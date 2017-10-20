#!/bin/bash
echo "Repeating command $* every second"
while sleep 1; do
    ./cli --host 192.168.86.28 --port 50060
done
