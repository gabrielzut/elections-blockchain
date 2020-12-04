"use strict";
const FabricCAServices = require("fabric-ca-client");
const { FileSystemWallet, X509WalletMixin } = require("fabric-network");
const fs = require("fs");
const path = require("path");

async function enrollAdmin(orgName, MSP) {
  try {
    const connectionProfilePathJSON = fs.readFileSync(
      path.resolve(
        __dirname,
        "..",
        "connections",
        `connection-${orgName}.json`
      ),
      "utf8"
    );

    const caInfo = JSON.parse(connectionProfilePathJSON).certificateAuthorities[
      `${orgName}CA`
    ];

    const ca = new FabricCAServices(
      caInfo.url,
      { verify: false },
      caInfo.caName
    );

    const wallet = new FileSystemWallet(
      path.join(process.cwd(), `../wallet/wallet-${orgName}`)
    );

    const adminExists = await wallet.exists("admin");
    if (adminExists) {
      console.error("Admin already exists");
      return;
    }

    const enrollment = await ca.enroll({
      enrollmentID: "admin",
      enrollmentSecret: "adminpw",
    });

    await wallet.import(
      "admin",
      X509WalletMixin.createIdentity(
        MSP,
        enrollment.certificate,
        enrollment.key.toBytes()
      )
    );
  } catch (e) {
    console.error(`Failed to enroll admin user "admin": ${e}`);
    process.exit(1);
  }
}

module.exports = { enrollAdmin };
