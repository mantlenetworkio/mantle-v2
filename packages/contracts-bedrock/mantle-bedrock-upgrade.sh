#!/bin/bash

network=$1
step=$2

if [ $network != "devnet" ] &&
  [ $network != "goerli" ] &&
  [ $network != "sepolia" ] &&
  [ $network != "mainnet" ]
then
  echo "invalid network, expected network: devnet, goerli, sepolia and mainnet"
  exit 0
fi

if [ $step -eq 0 ]
then
  echo "------------------------------------"
  echo "Deploy bedrock l1 contracts"
  echo "------------------------------------"
  rm deploy/*
  cp deploy-backup/bedrock-upgrade-deploy-scripts/step0/* deploy
  rm -rf deployments/mantle-${network}
  cp -r deployments/mantle-${network}-bak deployments/mantle-${network}
  yarn deploy --network mantle-${network}
fi


if [ $step -eq 1 ]
then
  echo "------------------------------------"
  echo "Do systemIndicator phase 1"
  echo "------------------------------------"
  rm deploy/*
  cp deploy-backup/bedrock-upgrade-deploy-scripts/step1/* deploy
  yarn deploy --network mantle-${network}
fi


if [ $step -eq 2 ]
then
  echo "------------------------------------"
  echo "update L2OutputOracleDynamicConfig in systemIndicator"
  echo "------------------------------------"
  rm deploy/*
  cp deploy-backup/bedrock-upgrade-deploy-scripts/step2/* deploy
  yarn deploy --network mantle-${network}
fi


if [ $step -eq 3 ]
then
  echo "------------------------------------"
  echo "Do systemIndicator phase 2"
  echo "------------------------------------"
  rm deploy/*
  cp deploy-backup/bedrock-upgrade-deploy-scripts/step3/* deploy
  yarn deploy --network mantle-${network}
fi



if [ $step -eq 4 ]
then
  echo "------------------------------------"
  echo "restore all deploy scripts"
  echo "------------------------------------"
  rm deploy/*
  git restore --stage deploy/
  git restore deploy/
fi
