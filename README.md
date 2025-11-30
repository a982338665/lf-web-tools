

本地调试：
cd ./gin-web-server && go run main.go


git bash执行：
cd ./gin-web-server && ./build.sh


windows-cmd执行：
cd gin-web-server; ./build.bat
或者
cd gin-web-server; & ./build.bat
cmd /c "cd gin-web-server && ./build.bat"


powershell执行：
cd gin-web-server
if ($?) { ./build.bat }
   

服务器启动：
nohup ./app-linux-amd64 >> mss.log  2>& 1 &


本地启动
cd gin-web-server
#go mod tidy
go run main.go

