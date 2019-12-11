#!/bin/sh

gocov convert $1 | gocov-html
