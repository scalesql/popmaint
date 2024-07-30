Defrag
======
* I read through an FQDN and get back a list of writeable databases
* This gets converted to
    * FQDN -> Listener[0] if in AG
    * Database
* In catch up loop
    * connect to the Listener instead of the original FQDN
    * delete it if we can't reach it
