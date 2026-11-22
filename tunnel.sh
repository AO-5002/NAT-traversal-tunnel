#!/bin/bash

# Variables

ENV_FILE="./internal/protocol/.env"
CREATE_ENV=true

# Check if the .env already exists.

if [[ -f "$ENV_FILE" ]]; then
  echo "It appears a .env configuration exists"
  # shellcheck disable=SC2162
  read -p "Would you like to overwrite it? (y/N): " ANSWER
  [[ "$ANSWER" != [yY] ]] && CREATE_ENV=false
fi

# Create .env file

if $CREATE_ENV; then
  # Prompt user for Relay's host-address & port
  read -p "Path to EC2 key (.pem): " KEY
  KEY="${KEY/#\~/$HOME}"
  read -p "Enter EC2 IPv4-address: " SSH_HOST
  read -p "Enter EC2 port (default 8080): " SSH_PORT
  read -p "Enter EC2 user (default ec2-user): " SSH_USER
  read -p "Enter content location URL: " LOCAL_URL
  SSH_PORT="${SSH_PORT:-8080}"
  SSH_USER="${SSH_USER:-ec2-user}"

  # .env properties
  touch "$ENV_FILE"
  echo "DEST_SSH_KEY=$KEY" > "$ENV_FILE"
  echo "DEST_IPV4_USER=$SSH_USER" >> $ENV_FILE
  echo "RELAY_SECRET=$(openssl rand -hex 32)" >> $ENV_FILE
  echo "DEST_IPV4_NETWORK=tcp" >> $ENV_FILE
  echo "DEST_IPV4_HOST=$SSH_HOST" >> $ENV_FILE
  echo "DEST_IPV4_PORT=$SSH_PORT" >> $ENV_FILE
  echo "LOCAL_URL=$LOCAL_URL" >> $ENV_FILE

  # Alter perms on the file
  chmod 600 $ENV_FILE
fi

# Send the .env to the EC2 instance
# Load config
source "$ENV_FILE"
EC2="$DEST_IPV4_USER@$DEST_IPV4_HOST"

# Send the secret to the EC2
grep '^RELAY_SECRET=' "$ENV_FILE" | ssh -i "$DEST_SSH_KEY" "$EC2" 'umask 077; mkdir -p ~/relay; cat > ~/relay/.env' || { echo "Failed to send secret"; exit 1; }

# Build the relay image and send it to the EC2
docker build --platform linux/arm64 -t relay . || exit 1
docker save relay | ssh -i "$DEST_SSH_KEY" "$EC2" 'docker load' || exit 1

# Start the relay (removes the old one first)
ssh -i "$DEST_SSH_KEY" "$EC2" 'docker rm -f relay 2>/dev/null; docker run -d --name relay --env-file ~/relay/.env -p 8080:8080 relay' || exit 1

# Run the client
go run ./cmd/client/main.go