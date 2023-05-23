### 本代码包主要实现更新cve数据库中description为中文
Usage:
  nvdbtools [command]

Available Commands:
  cnnvd       实现cnnvd数据相关操作
  completion  Generate the autocompletion script for the specified shell
  cve         实现scanner的cvedb数据库相关操作
  help        Help about any command

Flags:
  -h, --help   help for nvdbtools

Use "nvdbtools [command] --help" for more information about a command.


 1、（nvdbtools cnnvd getxml）从cnnvd官网下载发部数据文件(xml)，并做预处理。   
 2、（nvdbtools cnnvd importDB）将下载的xml文件导入到sqlite数据库。   
 3、getdbfile.sh脚本从 scanner镜像中提取cvedb数据库文件到目录：/tmp/nvdbtools/   
 4、（nvdbtools cve unizp）从cvedb中提取文件到目录：/tmp/nvdbtools/cvedbsrc   
 5、（nvdbtools cve update）从第2步生产称的cnvd数据库中提取中文description，用于更新第4步的文件中的description，并将跟新后的文件保存到目录：/tmp/nvdbtools/cvedbtarget/   
 6、（nvdbtools cve rebuild）重新打包cvedb数据文件到目录：/tmp/nvdbtools/cvedbtemp/   

