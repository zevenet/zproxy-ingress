#!/bin/bash

BASEDIR=$(dirname "$0")

YAML="zproxy-ingress-deplyment.yaml"
EXAMPLESDIR="examples"

cd $BASEDIR
rm $YAML

for FILE in `ls ${EXAMPLESDIR}/0*`
do
  cat $FILE >> $YAML
  echo '---' >> $YAML
done
