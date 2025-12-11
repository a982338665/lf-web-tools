cd /data/webtools/
kill -9 `pgrep -f app-linux-amd64`
nohup ./app-linux-amd64 >> p2p.log  2>& 1 &
