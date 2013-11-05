#!/bin/bash

TEMP_DIR=mmgorogue-deploy
BIN_DIR=$GOPATH/bin
OUT_FILE=mmgorogue-deploy

mkdir $TEMP_DIR

cp $BIN_DIR/mmgorogue $TEMP_DIR/
cp -R world $TEMP_DIR/

rm -f $OUT_FILE.tgz
tar -cvzf $OUT_FILE.tgz $TEMP_DIR

rm -rf $TEMP_DIR
