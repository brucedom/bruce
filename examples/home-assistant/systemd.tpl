[Unit]
Description=Open source home automation that puts local control and privacy first.
Documentation=https://www.home-assistant.io/getting-started/
After=docker.service
Requires=docker.service

[Service]
User=ha
Group=ha
TimeoutStartSec=120
Restart=always
Environment="IMAGE=homeassistant/home-assistant:latest"
ExecStartPre=-/usr/bin/docker stop %n
ExecStartPre=/usr/bin/docker pull ${IMAGE}
ExecStart=/usr/bin/docker run --rm --net=host \
  -v /data/home-assistant:/config \
  -v /etc/localtime:/etc/localtime:ro \
  -p 8123:8123 \
  --name %n ${IMAGE}

[Install]
WantedBy=multi-user.target