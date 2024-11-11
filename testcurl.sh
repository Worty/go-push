#!/bin/bash
for i in {1..100}; do
    cat testfile.bin | curl -v -F "user=password" -F "d=@-;filename=bla.$i" http://localhost:8080/ -w "\n"
done

hyperfine --warmup 100 'cat testfile.bin | curl -o /dev/null -s -F "user=password" -F "d=@-;filename=bla.$i" http://localhost:8080/'
