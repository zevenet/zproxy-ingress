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


ListenHTTP
        Address #LISTENERIP#
        Port #HTTPPORT#
        xHTTP 4
        RewriteLocation 1
End
