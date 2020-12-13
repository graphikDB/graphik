# GraphikDB - An Identity Aware MultiModel Database & Pubsub Server

## Abstract

## Problem Statement

1) Traditional relational databases are powerful but come with a number of issues that interfere with agile development methodologies:

- [database schema](https://en.wikipedia.org/wiki/Database_schema) setup requires application context & configuration overhead
- database [schema changes are often dangerous](https://wikitech.wikimedia.org/wiki/Schema_changes#Dangers_of_schema_changes) and require skilled administration to pull off without downtime
- [password rotation](https://www.beyondtrust.com/resources/glossary/password-rotation) is burdensome and leads to password sharing/leaks
- user/application passwords are generally stored in an external store(identity provider) causing duplication of passwords(password-sprawl)
- traditional database indexing requires application context in addition to database administration 

Because of these reasons, proper database adminstration requires a skilled database adminstration team(DBA).

This is bad for the following reasons:

- hiring dba's is [costly](https://www.payscale.com/research/US/Job=Database_Administrator_(DBA)/Salary)
- dba's generally have little context of the API accessing the database(only the intricacies of the database itself)
- communication between developers & dba's is slow (meetings & JIRA tickets)


2) Traditional non-relational databases are non-relational  
- many times developers will utilize a NOSQL database in order to avoid the downsides of traditional relational databases
- this leads to relational APIs being built on non-relational databases

Because of these reasons, APIs often end up developing anti-patterns

-  references to related objects/records are embedded within the record itself instead of joined via foreign key
- references to related objects/records are stored as foreign keys & joined via multiple requests

This is bad for the following reasons:

- as an API scales, the number of relationships will often grow, causing embedded relationships and/or multi-request queries to grow 
- embedded relational documents cause the single document to continue to grow in size forever
- foreign key relations often require multi-request queries(slow!)

3) No Awareness of origin/end user accessing the records(only the server making the request)
- database "users" are generally expected to be database administrators and/or another API.
- 3rd party [SSO](https://en.wikipedia.org/wiki/Single_sign-on) integrations are generally non-native
- databases may be secured properly by dba team while the APIs connecting to them can be insecure depending on the end user
