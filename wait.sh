#!/bin/bash

# Loop until curl returns a success or timesout attempts.
counter=0
curl --fail --silent --show-error $1 > /dev/null
while [ $? -ne 0 ] && [ "$counter" -lt 60 ]; do
  counter=$((counter+1))
  echo "Waiting.";
  sleep 1;
  curl --fail --silent --show-error $1 > /dev/null;
done;
