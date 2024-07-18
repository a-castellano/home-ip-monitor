#!/bin/sh

echo "### NOT starting on installation, please execute the following statements to configure windmaker-home-ip-monitor to start automatically using systemd"
echo " sudo /bin/systemctl daemon-reload"
echo "### This service is executed using systemd timers"
echo "### Enable it with the following command"
echo " sudo /bin/systemctl enable windmaker-home-ip-monitor.timer"
