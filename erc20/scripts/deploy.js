// We require the Hardhat Runtime Environment explicitly here. This is optional
// but useful for running the script in a standalone fashion through `node <script>`.
//
// You can also run a script with `npx hardhat run <script>`. If you do that, Hardhat
// will compile your contracts, add the Hardhat Runtime Environment's members to the
// global scope, and execute the script.
// const hre = require("hardhat");

async function main() {
  const [deployer] = await ethers.getSigners(); //get the account to deploy the contract
  console.log("Deploying contracts with the account:", deployer.address, "balance:", (await deployer.getBalance()).toString()); 


  const USDTContract = await ethers.getContractFactory("USDT");
  const USDT = await USDTContract.deploy("0xE48a7F0d63D00b5c209CB663bac0ec3e1410f7b7");
  console.log("Contract deploying tx:", USDT.deployTransaction.hash)

  await USDT.deployed();
  console.log("USDT deployed to:", USDT.address);
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
