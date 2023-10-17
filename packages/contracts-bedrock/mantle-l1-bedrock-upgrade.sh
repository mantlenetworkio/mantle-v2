#!/bin/bash

a=$1

if [ $a -eq 0 ]
then
  echo "------------------------------------"
  echo "Deploy bedrock l1 contracts"
  echo "------------------------------------"
  rm deploy/*
  cp deploy-backup/bedrock-upgrade-deploy-scripts/step0/* deploy
  rm -rf deployments/devnetL1-mantle-bedrock-upgrade
  cp -r deployments/devnetL1-mantle-bedrock-upgrade-bak deployments/devnetL1-mantle-bedrock-upgrade
  yarn deploy --network devnetL1-mantle-bedrock-upgrade
fi


if [ $a -eq 1 ]
then
  echo "------------------------------------"
  echo "Do systemIndicator phase 1"
  echo "------------------------------------"
  rm deploy/*
  cp deploy-backup/bedrock-upgrade-deploy-scripts/step1/* deploy
  yarn deploy --network devnetL1-mantle-bedrock-upgrade
fi


if [ $a -eq 2 ]
then
  echo "------------------------------------"
  echo "update L2OutputOracleDynamicConfig in systemIndicator"
  echo "------------------------------------"
  rm deploy/*
  cp deploy-backup/bedrock-upgrade-deploy-scripts/step2/* deploy
  yarn deploy --network devnetL1-mantle-bedrock-upgrade
fi


if [ $a -eq 3 ]
then
  echo "------------------------------------"
  echo "Do systemIndicator phase 2"
  echo "------------------------------------"
  rm deploy/*
  cp deploy-backup/bedrock-upgrade-deploy-scripts/step3/* deploy
  yarn deploy --network devnetL1-mantle-bedrock-upgrade
fi



if [ $a -eq 4 ]
then
  echo "------------------------------------"
  echo "restore all deploy scripts"
  echo "------------------------------------"
  rm deploy/*
  git restore --stage deploy/
  git restore deploy/
fi
