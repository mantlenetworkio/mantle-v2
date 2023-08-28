#!/bin/bash

a=$1

if [ $a -eq 0 ]
then
  rm -rf deployments/devnetL1-mantle
  cp -r deployments/devnetL1-mantle-bak deployments/devnetL1-mantle
  yarn deploy --network devnetL1-mantle
fi


if [ $a -eq 1 ]
then
  echo $a
  mv deploy deploy-deployment
  mv deploy-upgrade-1 deploy
  yarn deploy --network devnetL1-mantle
fi


if [ $a -eq 2 ]
then
  echo $a
  mv deploy deploy-upgrade-1
  mv deploy-upgrade-2 deploy
  yarn deploy --network devnetL1-mantle
fi


if [ $a -eq 3 ]
then
  echo $a
  mv deploy deploy-upgrade-2
  mv deploy-upgrade-3 deploy
  yarn deploy --network devnetL1-mantle
fi



if [ $a -eq 4 ]
then
  echo $a
  mv deploy deploy-upgrade-3
  mv deploy-deployment deploy
fi
