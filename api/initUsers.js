const { registerUser } = require("./registerUser");
const fs = require("fs");
const path = require("path");

async function initUsers(orgName, MSP) {
  try {
    const votingZonesString = fs.readFileSync(
      path.join(__dirname, "votingZones.json"),
      "utf-8"
    );
    const votingZones = JSON.parse(votingZonesString);

    for (votingZone of votingZones.machines) {
      await registerUser(orgName, votingZone.code, MSP);
    }
  } catch (e) {
    console.error("Error initializing users: " + e);
  }
}

module.exports = { initUsers };
