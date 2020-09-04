# Init template

Daemon			0
LogLevel        #LOGSLEVEL#
LogFacility		-
Timeout         #TOTALTO#
ConnTO          #CONNTO#
Alive           #ALIVETO#
Client          #CLIENTTO#
Control         "#SOCKETFILE#"
DHParams		"#DHFILE#"
ECDHCurve		"#ECDHCURVE#"
Ignore100Continue #IGNORE100CONTINUE#


ListenHTTP
        Address #LISTENERIP#
        Port #HTTPPORT#
        xHTTP 4
        RewriteLocation 1
End
