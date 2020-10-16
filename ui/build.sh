#!/bin/sh
npm run-script build
statik -src=./build -f
