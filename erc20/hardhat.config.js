require("@nomicfoundation/hardhat-toolbox");
require('@openzeppelin/hardhat-upgrades');

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: "0.8.17",
  defaultNetwork: "goerli",
  networks: {
    goerli: {
      url: "https://eth-goerli.g.alchemy.com/v2/TOKEN",
      accounts: [
        "PRIVATE_KEY"
      ],
      confirmations: 1,
    }
  },
  etherscan: {
    apiKey: "SCAN_API_KEY"
  }
};
