#!/usr/bin/env bash
name=$(basename $(pwd))
git remote add origin https://github.com/ab36245/${name}.git
git branch -M main
git push -u origin main
