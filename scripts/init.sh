#!/bin/bash
# Adapted from Philipp's script: https://github.com/wmnnd/nginx-certbot/blob/master/init-letsencrypt.sh

export COMPOSE_FILE=$1
export APP_ENVIRONMENT=prod

domains=(nmyk.io www.nmyk.io nickmykins.com www.nickmykins.com tmpch.at s.tmpch.at)
rsa_key_size=4096
data_path="./data/certbot"
email="nick@nmyk.io"
staging=0

if ! [-x "$(command -v make)" ]; then
	echo "❯ Installing make ..."
	sudo apt install make -y
fi

if ! [ -x "$(command -v docker-compose)" ]; then
  echo "❯ Installing docker ..."
  DOCKER_PLATFORM=https://download.docker.com/linux/ubuntu
  curl -fsSL $DOCKER_PLATFORM/gpg | sudo apt-key add -
  sudo apt-key fingerprint 0EBFCD88
  sudo add-apt-repository \
     "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
     $(lsb_release -cs) \
     stable" -y 
  sudo apt-get update
  sudo apt install docker-ce docker-compose -y
fi

if [ -d "$data_path" ]; then
  read -p "❯❯❯ Existing data found for $domains. Continue and replace existing certificate? (y/N) " decision
  if [ "$decision" != "Y" ] && [ "$decision" != "y" ]; then
    exit
  fi
fi

if [ ! -e "$data_path/conf/options-ssl-nginx.conf" ] || [ ! -e "$data_path/conf/ssl-dhparams.pem" ]; then
  echo "❯ Downloading recommended TLS parameters ..."
  mkdir -p "$data_path/conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > "$data_path/conf/options-ssl-nginx.conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem > "$data_path/conf/ssl-dhparams.pem"
fi

echo "❯ Creating dummy certificate for $domains ..."
path="/etc/letsencrypt/live/$domains"
mkdir -p "$data_path/conf/live/$domains"
docker-compose run --rm --entrypoint "\
  openssl req -x509 -nodes -newkey rsa:1024 -days 1\
    -keyout '$path/privkey.pem' \
    -out '$path/fullchain.pem' \
    -subj '/CN=localhost'" certbot

echo "❯ Starting nginx ..."
docker-compose up --force-recreate -d nginx

echo "❯ Deleting dummy certificate for $domains ..."
docker-compose run --rm --entrypoint "\
  rm -Rf /etc/letsencrypt/live/$domains && \
  rm -Rf /etc/letsencrypt/archive/$domains && \
  rm -Rf /etc/letsencrypt/renewal/$domains.conf" certbot

echo "❯ Requesting Let's Encrypt certificate for $domains ..."
#Join $domains to -d args
domain_args=""
for domain in "${domains[@]}"; do
  domain_args="$domain_args -d $domain"
done

# Email flag
email_arg="--email $email"

# Enable staging mode if needed
if [ $staging != "0" ]; then staging_arg="--staging"; fi

docker-compose run --rm --entrypoint "\
  certbot certonly --webroot -w /var/www/certbot \
    $staging_arg \
    $email_arg \
    $domain_args \
    --rsa-key-size $rsa_key_size \
    --agree-tos \
    --force-renewal" certbot

echo "❯ Reloading nginx ..."
docker-compose exec nginx nginx -s reload
docker-compose down
echo "Done. Run docker-compose up to start."
