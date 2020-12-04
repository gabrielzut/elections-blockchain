const { FileSystemWallet, Gateway } = require("fabric-network");
const path = require("path");
const channel = "mychannel";

async function getContract(conn, orgName, contractName, user) {
  const wallet = new FileSystemWallet(
    path.join(process.cwd(), `../wallet/wallet-${orgName}`)
  );

  const userExists = await wallet.exists(user);

  if (!userExists) {
    console.error(`User "${user}" does not exist in the wallet`);
    return;
  }

  const gateway = new Gateway();
  await gateway.connect(path.resolve(__dirname, "..", "connections", conn), {
    wallet,
    identity: user,
    discovery: {
      enabled: true,
      asLocalhost: true,
    },
  });

  const network = await gateway.getNetwork(channel);

  return network.getContract(contractName);
}

module.exports = { getContract };
