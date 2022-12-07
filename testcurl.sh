#!/bin/bash
for i in {1..100}; do
    cat testfile.bin | curl -F "user=pw" -F "d=@-;filename=bla.$i" http://localhost:8080/ -w "\n"
done
