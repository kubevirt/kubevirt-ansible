#!/bin/sh
for num in `seq 30 50` ; do
 ip r | grep -q 172.$num
 if [ "$?" != "0" ] ; then
  printf 172.$num.0.0/16
  exit
 fi
done
printf 172.30.0.0/16
