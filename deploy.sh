#!/bin/bash

set -e

CDN_ASSETS_PATH=s3://nmyk/assets/
NGINX_ROOT=/var/www/html/

upload_to_spaces() {
	for file in "$@"; do
		s3cmd -c .nmyk.io.s3cfg put $file $CDN_ASSETS_PATH
		s3cmd -c .nmyk.io.s3cfg setacl $CDN_ASSETS_PATH$(basename $file) --acl-public
	done
}

upload_to_spaces assets/*
scp index.html root@nmyk.io:$NGINX_ROOT
echo "Done"
