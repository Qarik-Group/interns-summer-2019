##Introducing SHIELD etcd plugin

We spent the summer over at Stark&Wayne to design and build the etcd plugin for SHIELD. The basic functionalities include backing up data from etcd and then restoring this data to etcd using the backup archives. This blog post will explain all the nitty gritty details about the functionalities that the new target plugin supports.

Etcd is an open-source distributed key-value store that serves as the backbone of distributed systems by providing a canonical hub for cluster coordination and state management – the systems source of truth. While etcd was built specifically for clusters running CoreOS, etcd works on a variety of operating systems including OS X, Linux, and BSD.

Kubernetes lives on top of etcd. Etcd’s job within Kubernetes is to safely store critical data for distributed systems. It’s best known as Kubernetes’ primary datastore used to store its configuration data, state, and metadata. Since Kubernetes usually runs on a cluster of several machines, it is a distributed system that requires a distributed datastore like Etcd. In essence, the data in etcd is all that an operator needs to resurrect a dead cluster.

As a part of this project's requirement, we validated both etcd and SHIELD. Like all other SHIELD plugins, this plugin is also written in Go.


##How does the plugin work?

Backups and restores are the backbone of the SHIELD contract. The etcd plugin works along the same principles that govern the SHIELD realm. We have extensivley leveraged the `etcd/clientv3` for writing the plugin; which is the official Go etcd client for v3. The etcd plugin supports both single node and multi node setups. 

###Authentication

etcd plugin supports:

1. Anonymous auth(no auth)
2. Role-based auth (Username/password)
3. Certificate based auth
4. Role based auth coupled with cert based auth

###Fields

This is the defacto information that you need to provide any plugin in order to backup/restore. etcd plugin needs the following information:

1. Endpoints/etcd URLs: IP addresses/DNS names of your etcd installation
2. Timeout: Defaults to 2 seconds. 
3. Auth type: No auth, RBAC, cert based
4. Username
5. Password
6. Client certificate (User/Operator needs to copy paste the contents of the PEM file)
7. Client key (User/Operator needs to copy paste the contents of the PEM file)
8. CA cert (User/Operator needs to copy paste the contents of the PEM file)
9. Overwrite mode: This mode enables deletion of exisiting state of the etcd cluster. Restores the cluster with an earlier sane backup. Super helpful in case of corrupted keys/values.
10. Prefixed backup: You can specify a directory/key_name structure to backup only certain keys with the same prefix.

###Validation

etcd plugin duly validates all the fields that the User/Operator provides after the connection to the etcd installation is established. Checks all the dependencies and moves to backup the cluster once everything gets verified.

###Backup

After validation, etcd plugin moves forward to backup the installation and dumps the keys and respective values to a store plugin of your choice (SHIELD supported storage plugins). You can configure your backups to be periodic/automated as well. 

###Restore

Once the backup job is run, you can use the etcd plugin to retrieve the backups and use them to restore your installations. SHEILD will pull in the archives from the storage and restore them.

###Error Handling

etcd plugin returns following types of errors:

1. context error: canceled or deadline exceeded.
2. Keys/values decoding errors.
3. etcd cluster/installation's invalid configuration errors.




##How to try out the etcd plugin?

If you have an etcd installation (single node/multi-node) setup somewhere, we highly suggest getting the latest SHIELD release (v8.4.0) to test out the shiny new plugin. You can do that using a bunch of options:

1. Download SHIELD via  [https://github.com/shieldproject/shield/releases/tag/v8.4.0]()
2. `go get github.com/shieldproject/shield`. Just make sure your GOPATH is set correctly. This can be a huge pain otherwise. 
3. Docker compose recipe for this release will be soon available. Details will be updated here.

We hope that you'll have fun experimenting with the etcd plugin. Let us know how this goes. 

