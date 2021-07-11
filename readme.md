go mod 环境配置
go env 开启GO111MODOULE
先go init创建go mod init 后面跟着本项目在仓库的路径 如 go init github.com/Jundong-Chan/seckill 命令要在项目文件夹下执行
然后go proxy 要配置成能用的，https://goproxy.cn,direct
然后要在goland 中配置go moudule perference---go---gomodoule  配置环境中的GOMODCACHE
go mod tidy 会自动下载依赖包
go mod build 会自动在go.mod文件中生成依赖
