# Goroxy
Simple proxy with golang. No root, port forwarding, other dependency or library needed, just pure golang.

Support http 1.1 </br>
Support chain proxy(use 'client' on not-final servers)</br>
Cross platform(Windows, Linux, MacOS, Android)</br>
Open source <a href="https://github.com/khordady/Goroxy_Android">Android app</a></br>
AES/ECB encryption for innitial connection</br>
Use tcp connection, so will never be blocked</br>
Use user/pass for protection</br>

<h3>how to use</h3>
1) Download <a href="https://github.com/khordady/goroxy/releases">Relaese files</a></br>
2) config server and client .json file</br>
3) run Server on your vps</br>
4) run Client in your local pc</br>
5) set http/https proxy in your app (like firefox) and coonect to Client</br>
hope you enjoy :)</br>

<h3>how to compile and use on any os</h3>
1) Download and install golang sdk atleast 1.19</br>
2) compile each one seperatly</br>
3) config .json file</br>
4) and then run</br>
hope you enjoy  :)</br>
follow <a href="https://golangdocs.com/install-go-linux">this link</a> for install golang in linux

<h3>#Tips</h3>
1)Always use AES with 32 char</br>
2)Always use Strong User/Pass</br>
3)don't use for any bad ...hub or you will burn in hell :|</br>

sample config file(Case sensitive!):</br>

{</br>
<--this part is for chain proxy</br>
  "ListenPort": "8000",</br>
  "ListenEncryption": "None", //or AES</br>
  "ListenEncryptionKey": "SOMETHING 16 bit", //or 24 or 32 en character</br>
  "ListenAuthentication": false,</br>
  "ListenUsers": [</br>
    {</br>
      "ListenUserName": "Goroxy",</br>
      "ListenPassword": "Goroxy"</br>
    },</br>
    {</br>
      "ListenUserName": "Goroxy2",</br>
      "ListenPassword": "Goroxy2"</br>
    }</br>
  ],</br>
  --></br>
  "Server": "192.168.1.101",</br>
  "ServerPort": "8181",</br>
  "SendEncryption": "AES",  //or None</br>
  "SendEncryptionKey": "SOMETHING 16 bit", //or 24 or 32</br>
  "SendAuthentication": true,</br>
  "SendUserName": "Goroxy",</br>
  "SendPassword": "Goroxy"</br>
}</br>
