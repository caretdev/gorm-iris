#!/bin/bash

for test in ../../gorm/tests/*_test.go; do
  file=$(basename $test)
  if [ ! -f $file ]; then
    if [ $(grep 'DB' $test | wc -l) -eq 0 ]; then
      continue
    fi
    if [ $(grep -i 'lib/pq' $test | wc -l) -gt 0 ]; then
      continue
    fi
    if [ $(grep -i 'gauss' $test | wc -l) -gt 0 ]; then
      continue
    fi
    if [ $(grep -i 'mysql' $test | wc -l) -gt 0 ]; then
      continue
    fi
    echo $test
    # cp $test .
    # break
  fi
done