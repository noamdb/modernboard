#!/bin/bash

# read user input
read -p 'Username: ' user
read -sp 'Password: ' pass
echo 
read -p 'App name (name your application):' appName
read -p 'API domain (ie: api.modernboard.com): ' apiDomain
read -p 'Website domain (ie: https://modernboard.com): ' websiteDomain

repoPath=$(git rev-parse --show-toplevel)
cd /tmp # change directory to a directory that everyone has permissions on


# add user with sudo
adduser --disabled-password --gecos "" $user
echo ${user}:${pass} | chpasswd	
usermod -aG sudo ${user}

# configure firewall
sudo ufw allow OpenSSH
sudo ufw allow proto tcp from any to any port 80,443

# install and configure postgres
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -sc)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
sudo apt update
sudo apt install postgresql-11 -y
sudo systemctl enable postgresql.service
sudo systemctl start postgresql.service
sudo -u postgres createuser -s ${user}
sudo -u postgres psql -c "alter user ${user} with encrypted password '${pass}';"
sudo -u postgres createdb $user


# install go
sudo snap install go --classic


# install graphicsmagick
sudo apt install build-essential software-properties-common libjpeg-dev libtiff5-dev libpng16-dev --fix-missing -y
wget https://sourceforge.net/projects/graphicsmagick/files/graphicsmagick/1.3.31/GraphicsMagick-1.3.31.tar.gz
tar -xvf GraphicsMagick-1.3.31.tar.gz
cd GraphicsMagick-1.3.31
./configure
make
make install
rm GraphicsMagick-1.3.31.tar.gz GraphicsMagick-1.3.31

# install ffmpeg
sudo add-apt-repository ppa:jonathonf/ffmpeg-4 -y
sudo apt update
sudo apt install ffmpeg -y

# create the database tables
cp ${repoPath}/create.sql /tmp/create.sql
sudo -u postgres psql -d ${user} -f /tmp/create.sql
rm /tmp/create.sql

# build the application
cd ${repoPath}
go build
git update-index --assume-unchanged ./config/conf.yaml
salt=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 13 ; echo '')
cat >./config/conf.yaml <<EOL
admin_name: "admin"
admin_password: "admin"
domain: "${apiDomain}"
cors_domains: ["${websiteDomain}"]
static_path: "/home/${user}/static/"
environment: "production"
cert_dir: "secret-dir"
database_url: "postgres://${user}:${pass}@localhost:5432/${user}?sslmode=disable"
default_salt: "${salt}"
max_image_size_mb: 10
EOL
mkdir -p /home/${user}/static/{files,thumbnails}
./modernboard -init=true
cd /tmp





# give executable permission to listen on ports
sudo setcap CAP_NET_BIND_SERVICE=+eip ${repoPath}/modernboard

sudo touch /var/log/${appName}.log
 
cat >/etc/rsyslog.d/${appName}.conf <<EOL
if \$programname == '${appName}' then /var/log/${appName}.log
& stop
EOL
sudo chown syslog /var/log/${appName}.log
sudo systemctl restart rsyslog

# configure the service
cat >/etc/systemd/system/${appName}.service <<EOL
[Unit]
Description = ${appName} api service
After = network.target
[Service]
WorkingDirectory = ${repoPath}
ExecStart = ${repoPath}/modernboard
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=${appName}
[Install]
WantedBy = multi-user.target
EOL

sudo systemctl daemon-reload
systemctl enable ${appName}.service
systemctl start ${appName}.service