#!/bin/bash
YEAR=$(date +%Y)
LAST_VERSION_TAG=$(git tag -l --sort=-creatordate|grep -v 'pre'|head -n 1)
LAST_MAJOR=$(echo $LAST_VERSION_TAG|sed -E 's/^[^\.]+\.//g;s/\..+//g')
MAJOR=$(($LAST_MAJOR+1))
MINOR=$(git log --oneline HEAD...$LAST_VERSION_TAG|wc -l)
echo "Version: "$YEAR.$MAJOR.$MINOR