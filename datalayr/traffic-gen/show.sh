#!/bin/bash

log_dir=$1

for entry in $(ls $log_dir); do 
	tail -n 1 $log_dir/$entry
done 
