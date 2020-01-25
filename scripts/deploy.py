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

with open('secrets.env', 'r') as secrets_file:
    secrets = dict()
    for line in secrets_file:
        k, v = line.split('=')
        secrets[k] = v.strip()

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


def with_secrets_on_remote_host(func):
    def wrapped_func(*args, **kwargs):
        os.system('scp secrets.env root@nmyk.io:~/nmyk.io')
        func(*args, **kwargs)
        os.system('ssh root@nmyk.io "rm ~/nmyk.io/secrets.env"')
    return wrapped_func


@with_secrets_on_remote_host
def restart_app():
    cmd = 'ssh root@nmyk.io /bin/bash'.split()
    heredoc = ("<<-'ENDSSH'\n"
               "cd nmyk.io\n"
               "make stop\n"
               "git checkout -- . && git pull\n"
               "make run-prod\n"
               "ENDSSH")
    cmd.append(heredoc)
    if subprocess.run(cmd).returncode:
        print('Could not restart app')
        sys.exit(1)
    print('❯ https://nmyk.io')


def main():
    for asset in os.listdir('web/assets'):
        upload(asset)
    restart_app()
    print('Done')


if __name__ == '__main__':
    main()
