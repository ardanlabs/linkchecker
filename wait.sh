#!/bin/bash

# Loop until curl returns a success or 30 attempts.
counter=0
curl --fail --silent --show-error $1 > /dev/null
while [ $? -ne 0 ] && [ "$counter" -lt 30 ]; do
  counter=$((counter+1))
  echo "Waiting.";
  sleep 1;
  curl --fail --silent --show-error $1 > /dev/null;
done;
