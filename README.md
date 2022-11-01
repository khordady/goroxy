# Goroxy
Simple proxy with golang. No root, port forwarding, other dependency or library needed, just pure golang.

Support http 1.1 
Support chain proxy(use 'client' on not-final servers)
Cross platform(Windows, Linux, MacOS, Android)
Open source <a href="https://github.com/khordady/Goroxy_Android">Android app</a>
AES/ECB encryption for innitial connection
Use tcp connection, so will never be blocked
Use user/pass for protection
just doesn't wash your dishes :(

how to use
1) Download <a href="https://github.com/khordady/goroxy/releases">Relaese files</a>
2) config server and client .json file
3) run Server on your vps
4) run Client in your local pc
5) set http/https proxy in your app (like firefox) and coonect to Client
hope you enjoy :)

how to compile and use on any os
1) Download and install golang sdk atleast 1.19
2) compile each one seperatly
3) config .json file
4) and then run
hope you enjoy  :)
follow <a href="https://golangdocs.com/install-go-linux">this link</a> for install golang in linux

#Tips
1)Always use AES with 32 char
1)Always use Strong User/Pass
2)use port 443 on vps(server) for anonymousity(or something like that :)
3)don't use for any bad ...hub or you will burn in hell :|

sample config file(Case sensitive!):

{
<--this part is for chain proxy
  "ListenPort": "8000",
  "ListenEncryption": "None", //or AES
  "ListenEncryptionKey": "SOMETHING 16 bit", //or 24 or 32 en character
  "ListenAuthentication": false,
  "ListenUsers": [
    {
      "ListenUserName": "Goroxy",
      "ListenPassword": "Goroxy"
    },
    {
      "ListenUserName": "Goroxy2",
      "ListenPassword": "Goroxy2"
    }
  ],
  -->
  "Server": "192.168.1.101",
  "ServerPort": "8181",
  "SendEncryption": "AES",  //or None
  "SendEncryptionKey": "SOMETHING 16 bit", //or 24 or 32
  "SendAuthentication": true,
  "SendUserName": "Goroxy",
  "SendPassword": "Goroxy"
}
