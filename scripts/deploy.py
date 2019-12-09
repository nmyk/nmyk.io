#!/usr/local/bin/python3.6
import json
import os
import subprocess
import sys

import boto3

SPACES_REGION = 'nyc3'
SPACES_BUCKET = 'nmyk'

content_types = {
    '.css': 'text/css',
    '.js': 'text/javascript'
}

with open('secrets.json') as secrets_file:
    secrets = json.load(secrets_file)

spaces = boto3.client(
    's3',
    region_name=SPACES_REGION,
    endpoint_url=f'https://{SPACES_REGION}.digitaloceanspaces.com',
    aws_access_key_id=secrets['SPACES_KEY'],
    aws_secret_access_key=secrets['SPACES_SECRET']
)


def upload(asset):
    path = f'web/assets/{asset}'
    _, file_ext = os.path.splitext(path)
    with open(path, 'rb') as asset_file:
        resp = spaces.put_object(
            Bucket=SPACES_BUCKET,
            Key=f'assets/{asset}',
            Body=asset_file,
            ContentType=content_types.get(file_ext, 'binary/octet-stream'),
            ACL='public-read'
        )
    if not resp['ResponseMetadata']['HTTPStatusCode'] == 200:
        print(f'Could not upload to spaces: {asset}', resp)
        sys.exit(1)
    print(f'❯ {asset}')


def restart_app(compose_file):
    cmd = 'ssh root@nmyk.io /bin/bash'.split()
    heredoc = ("<<-'ENDSSH'\n"
               "cd nmyk.io\n"
               "make stop\n"
               "git checkout -- . && git pull\n"
               f"make run-prod\n"
               "ENDSSH")
    cmd.append(heredoc)
    if subprocess.run(cmd).returncode:
        print('Could not restart app')
        sys.exit(1)
    print('❯ https://nmyk.io')


def main():
    compose_file = sys.argv[1]
    for asset in os.listdir('web/assets'):
        upload(asset)
    restart_app(compose_file)
    print('Done')


if __name__ == '__main__':
    main()