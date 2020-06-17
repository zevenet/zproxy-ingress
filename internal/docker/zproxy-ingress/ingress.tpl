LogLevel        #LOGSLEVEL#
Timeout         45
ConnTO          20
Alive           10
Client          30
Control         "#SOCKETFILE#"

ListenHTTP
        Address #DEFAULTIP#
        Port #DEFAULTPORT#
        xHTTP 4
        RewriteLocation 1
End
