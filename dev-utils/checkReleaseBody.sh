#!/bin/bash
sed -E '/---/q;/^## .*$/d' ../CHANGELOG.md|sed '/^---$/d'|less