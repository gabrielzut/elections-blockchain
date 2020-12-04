const express = require("express");
const bodyParser = require("body-parser");
const crypto = require("crypto");
const app = express();
const { getContract } = require("./getContract");
const { enrollAdmin } = require("./enrollAdmin");
const { initUsers } = require("./initUsers");
app.set("view engine", "ejs");
app.use(bodyParser.json());
urlencoder = bodyParser.urlencoded({ extended: true });
const orgName = "Org1";
const orgMSP = "Org1MSP";
const conn = "connection-Org1.json";

const electionEndPeriod = process.argv.length >= 3 ? process.argv[2] : 86400000;

startElection();

async function startElection() {
  try {
    console.info("Enrolling admin user...");
    await enrollAdmin(orgName, orgMSP);
    console.info("Initializing users for the voting machines...");
    await initUsers(orgName, orgMSP);
    await initElection();

    setTimeout(() => {
      endElection();
    }, electionEndPeriod);
  } catch (error) {
    console.log(`Error starting election: ${error}`);
  }
}

async function initElection() {
  console.info("Initializing election...");

  const electionsContract = await getContract(
    conn,
    orgName,
    "elections",
    "admin"
  );

  const electionsLogContract = await getContract(
    conn,
    orgName,
    "elections_log",
    "admin"
  );

  await electionsContract.submitTransaction("initElection");
  await electionsLogContract.submitTransaction("initElection");

  console.info("Election initialized successfully!");
}

async function endElection() {
  console.info("Ending election...");
  try {
    const electionsContract = await getContract(
      conn,
      orgName,
      "elections",
      "admin"
    );

    const electionsLogContract = await getContract(
      conn,
      orgName,
      "elections_log",
      "admin"
    );

    await electionsContract.submitTransaction("endElection");
    await electionsLogContract.submitTransaction("endElection");

    console.info("Election ended successfully!");
  } catch (e) {
    console.error(`There was an error ending election: ${e}`);
  }
}

app.post("/api/vote", urlencoder, async function (req, res) {
  try {
    console.info(`Registering vote of voter ${req.body.voterId}...`);

    const electionsContract = await getContract(
      conn,
      orgName,
      "elections",
      req.body.zoneCode
    );

    const electionsLogContract = await getContract(
      conn,
      orgName,
      "elections_log",
      req.body.zoneCode
    );

    const confirmationKey =
      Math.random().toString(36).substring(2, 15) +
      Math.random().toString(36).substring(2, 15);

    await electionsLogContract.submitTransaction(
      "register",
      req.body.voterId,
      confirmationKey,
      new Date().toISOString()
    );

    const id = crypto
      .scryptSync(req.body.password, confirmationKey, 64)
      .toString("utf-8");

    await electionsContract.submitTransaction(
      "vote",
      id,
      req.body.candidateNumber,
      new Date().toISOString()
    );

    res.json({
      status: "Success",
    });

    console.info("Vote registered successfull!");
  } catch (e) {
    console.error(`Failed to register transaction: ${e.message}`);
    res.status(500).json({
      error: e.message,
      status:
        "Error registering transaction, please make sure the voter hasn't voted yet.",
    });
  }
});

app.post("/api/audit", async function (req, res) {
  try {
    console.info(`Auditing vote of ${req.body.voterId}...`);

    const electionsContract = await getContract(
      conn,
      orgName,
      "elections",
      req.body.zoneCode
    );

    const electionsLogContract = await getContract(
      conn,
      orgName,
      "elections_log",
      req.body.zoneCode
    );

    const electionLogBuffer = await electionsLogContract.evaluateTransaction(
      "getByVoterId",
      req.body.voterId
    );

    const electionLog = JSON.parse(electionLogBuffer.toString("utf-8"));

    if (electionLog.confirmationKey) {
      const electionVoteBuffer = await electionsContract.evaluateTransaction(
        "auditById",
        crypto
          .scryptSync(req.body.password, electionLog.confirmationKey, 64)
          .toString()
      );

      const electionVote = JSON.parse(electionVoteBuffer.toString("utf-8"));

      res.json({
        vote: {
          candidateNumber: electionVote.candidateNumber,
          dateTime: electionVote.dateTime,
        },
        status: "Success",
      });

      console.info("Vote audited successfully!");
    } else {
      res.json({ status: "Voter not found" });

      console.error("Failed to audit vote: voter not found!");
    }
  } catch (e) {
    console.error(`Failed to audit vote: ${e.message}`);
    res.status(500).json({
      error: e.message,
      status: "Error auditing vote",
    });
  }
});

app.get("/api/auditAll", async function (req, res) {
  try {
    console.info("Auditing all votes...");

    const electionsContract = await getContract(
      conn,
      orgName,
      "elections",
      req.body.zoneCode
    );

    const votesBuffer = await electionsContract.evaluateTransaction(
      "auditByRange",
      "",
      ""
    );

    const votes = JSON.parse(votesBuffer.toString("utf-8"));
    const votesByCandidate = {};

    if (votes) {
      for (const vote of votes) {
        if (vote.candidateNumber) {
          if (votesByCandidate[vote.candidateNumber]) {
            votesByCandidate[vote.candidateNumber] += 1;
          } else {
            votesByCandidate[vote.candidateNumber] = 1;
          }
        }
      }
    }

    res.json({ votes: votesByCandidate, status: "Success" });

    console.info("Votes audited successfully!");
  } catch (e) {
    console.error(`Failed to list votes: ${e.message}`);
    res.status(500).json({
      error: e.message,
      status: "Error listing votes",
    });
  }
});

app.listen(3000, () => {
  console.log("***********************************");
  console.log("API server listening at localhost:3000");
  console.log("***********************************");
});
