"use strict";

const {
  FileSystemWallet,
  Gateway,
  X509WalletMixin,
} = require("fabric-network");
const path = require("path");

async function registerUser(orgName, user, MSP) {
  try {
    const connectionProfilePath = path.resolve(
      __dirname,
      "..",
      "connections",
      `connection-${orgName}.json`
    );
    const walletPath = path.join(
      process.cwd(),
      `../wallet/wallet-${orgName}`
    );
    const wallet = new FileSystemWallet(walletPath);

    const userExists = await wallet.exists(user);

    if (userExists) {
      console.error(`User ${user} already exists`);
      return false;
    }

    const adminExists = await wallet.exists("admin");

    if (!adminExists) {
      console.error("Admin not found");
      return false;
    }

    const gateway = new Gateway();
    await gateway.connect(connectionProfilePath, {
      wallet,
      identity: "admin",
      discovery: { enabled: true, asLocalhost: true },
    });

    const ca = gateway.getClient().getCertificateAuthority();
    const adminIdentity = gateway.getCurrentIdentity();

    const secret = await ca.register(
      { enrollmentID: user, role: "client" },
      adminIdentity
    );
    const enrollment = await ca.enroll({
      enrollmentID: user,
      enrollmentSecret: secret,
    });
    const userIdentity = X509WalletMixin.createIdentity(
      MSP,
      enrollment.certificate,
      enrollment.key.toBytes()
    );

    await wallet.import(user, userIdentity);
  } catch (e) {
    console.error(`Failed to register user "${user}": ${e}`);
    process.exit(1);
  }
}

module.exports = { registerUser };
