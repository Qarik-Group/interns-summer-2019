The Design of the SHIELD Archive Encryption System
==================================================

The goals of the archive encryption system are:

  1. Archive encryption happens as early as possible
  2. Archive decryption happens as late as possible
  3. The same data does not produce the same encrypted archive
  4. Ciphers and modes can be upgraded by operators
  5. Operators are minimally involved in day-to-day management
     of encryption keys (i.e. Stuff Just Works)


Encryption Keys Management
--------------------------

For SHIELD to be able to handle the encryption and decryption of
archives, it needs to be able to generate, store, and retrieve
encryption keys and initialization vectors (IVs).

To facilitate this, SHIELD will contain a managed Vault instance,
single-threaded, backed by the same PostgreSQL storage as the main
SHIELD database.  This Vault will reside on loopback only, not
accessible from outside the core daemon installation.  Only the
SHIELD supervisor will communicate with the Vault, and it will be
responsible for initializing and unsealing the Vault.

This setup leaves the seal key and the initial root token to be
managed by and available to the SHIELD supervisor daemon.
These credentials will be encrypted with a single "master
password" that needs to be supplied to the supervisor daemon as
part of its startup routine.  The encrypted credentials can then
be left on-disk.

Whenever the API/UI detects that the backend Vault has been
sealed, it will enter into a "hibernation" mode where it does not
perform backup / recovery operations, locks out all of the API
endpoints, and displays a very large "unlock me" UI page.  When
an administrator provides the master password, either via the UI
page, or via `shield unlock` on the command-line.

### Consideration - Master Password Change

SHIELD administrators will need to be able to rotate the master
password at-will, once the Vault has been unsealed.  This
dovetails with the RBAC work that's on the roadmap, since only
adminsitrators should be granted that privilege.  Changing the
master password will decrypt the credentials using the in-force
password, re-encrypt them with the new master password and store
the new ciphertext.

### Consideration - Vault Initialization

Brand new SHIELD installations will start out life with no
database entries and a blank Vault.  The supervisor API will be
responsible for initializing the Vault to establish a single seal
key and generate an initial root token, but in order to encrypt
those credentials, administrator input is required (the master
password).

The supervisor API therefore needs to track the initialization
state of the Vault, and if the Vault is uninitialized, the web
interface should also enter a "hibernation" mode, much like a
sealed Vault, but instead of asking for the master password to
unseal the Vault, it asks for the master password in order to
initialize the Vault.

### Consideration - Monitoring

The SHIELD monitoring endpoint will need to track and export
details on current seal-status of the Vault.  If the Vault is
sealed for some reason (new boot, service restart, etc.) then the
monitoring system should be able to detect that so that
administrators are informed that backup jobs are effectively
paused, and not running.

Uninitialized Vaults should also ping on the monitoring system
interface.  In the normal case, these SHIELDs will not be in
monitoring yet, because often the Vault is only uninitialized
until the SHIELD has been fully deployed.  In the unusual and
as-yet-unseen case of a Vault being uninitialized after legitimate
work has been done, having monitoring go off is preferable to the
alternative.

### Consideration - Backing Up SHIELD Itself

One of the nice features of SHIELD being a backup system is that
it is able to quite easily backup its own configuration.  Once we
add encryption into the mix, however, it becomes a bit trickier.
We cannot just generate new key material for every backup of
SHIELD, since that key material will not be known to the SHIELD
site administrators during restore.

Instead, when a SHIELD backup is performed, instead of generating
new key material, the master password will be used, and an
initialization vector derived from it.  That way, during restore,
the administrator need only provide the master password that was
in-force when the backup archive was taken, in order to decrypt.

The mechanism for detecting SHIELD backups is simple: if the
target plugin is "shield", use the master password; otherwise,
generate new key material.

Configuration of Encryption and Generation of Key Material
----------------------------------------------------------

The SHIELD supervisor will be responsible for generating key
material (secret key and initialization vector), which it will
store in the Vault.  The key sizes will derive from the Encryption
Type, which is selected by the user deploying the SHIELD.

Changing the Encryption Type must be possible, to allow for
stronger future ciphers and chaining modes to be patched into the
binary code and activated by operators.  Doing so must not
_ransom_ pre-existing backups that were encrypted with different
ciphers and modes.

A new key + iv pair will be generated for every new archive (with
regrettable but necessary exception).  Because of this aggresive
key rotation schedule, key management is not affected by a change
in Encryption Type -- the next archive after such a change will
generate a key + iv pair for the new Encryption Type.

These three pieces of information must be stored for every archive
generated by the system: (1) Encryption Type, (2) Encryption Key,
and (3) Initialization Vector.

Encryption Types are encoded as the Encryption Cipher, followed by
a hyphen, and then the Encryption Mode.  The following are
currently valid, and will be documented in operator manuals:

**Ciphers**

  - aes128 - Advanced Encryption Standard with 128-bit keys.
  - aes256 - Advanced Encryption Standard with 256-bit keys.
  - twofish - Schneier's successor to blowfish, 256-bit keys.

**Modes**

  - cfb - Cipher Feedback Mode
  - ofb - Output Feedback Mode
  - ctr - Counter Mode

Thus, to use the Twofish block cipher in counter mode, the
Encryption Type is "twofish-ctr".  This string will be stored in
the archive record in the SHIELD database.

That leaves the Encryption Key and the Initialization Vector.
These will be stored in the Vault, because they are sensitive key
material.  These are both multi-byte quantities, and for durable
storage and diagnostic purposes, will be stored in ASCII
hexadecimal encoding, in grouped sets of 2 octets (4 hex digits):

    deca-fbad-abad-1dea

Each 2-octet / 4-digit group will be separated by a hyphen.

These encoded values will be stored in the Vault, at a path that
allows for easy retrieval later during download / restore
operations:

    secret/archives/<archive-uuid>:key
    secret/archives/<archive-uuid>:iv

Two other (non-sensitive) values will be stored alongisde these:

    secret/archvies/<archive-uuid>:uuid
    secret/archives/<archive-uuid>:type

`uuid` contains the archive UUID, and `type` identifies the
Encryption Type.  These are redundant, and exist
as a safety precaution to detect any accidental corruption of the
Vault -- eventually operators will put tools like `safe` on SHIELD
and go mucking about.

During deciphering activities, the supervisor will retrieve these
four values from the Vault, and then validate that the UUID
matches the archive UUID, and that the Encryption Type in the
Vault matches that in the Archive record.  If either of these
checks fail, the superviros will refuse to continue.  It is up to
the operator to manually intervene at this point and fix what's
wrong.


Propagation of Encryption Keys Throughout The System
----------------------------------------------------

The SHIELD Supervisor will need to convey encryption key material
(key + iv) to various parts of the system in order for it to
function.  This section attempts to delineate the ultimate extant
to which such material is conveyed.

All communication between the supervisor and the Vault itself
occurs inside the memory spaces of the `shieldd` and `vault`
processes, and over a secured and verified HTTPS channel 

The supervisor communicates the Encryption Type, Key and IV to the
chosen worker on the remote SHIELD agent via the SSH-secured
worker communication channel.  At that point, these credentials
enter the memory space of the SHIELD agent, possibly on a
different host.

The SHIELD agent process places the Encryption Type, Key and IV
in the environment of the `shield-pipe` orchestatrion process,
which will actually perform the encryption (on store) and
decryption (on retrieve) of the archive data.  These environment
variables _MUST NOT_ be propagated to the plugins themselves, to
limit exposure of the credentials.

### Consideration - TLS Certificate Validation

To protect communication between the SHIELD supervisor daemon and
the Vault, all transactions will be carried out over HTTPS/TLS.
For this to work, the implementation must be unyielding in
verifying certificates.

To that end, the SHIELD supervisor must be given a CA certificate
that it will use to validate the Vault certificate.  The
deployment of the SHIELD will be responsible for putting the
necessary CA certificates and CA-signed certificates where SHIELD
and Vault can find them.

### Consideration - Mitigating MMU/Swap with mlock

As key material propagates throughout the system, it will reside
in memory.  The longer it resides in memory, the greater the
chance that the operating system kernel will swap the containing
page to disk, where it will remain until an indeterminate point in
the future.

This presents another vector for credentials leakage, one that we
may be able to mitigate by judicious use of _memory locking_ and
the `mlock(2)` system call.

Go's `golang.org/x/sys/unix` exposes the `Mlock()` call on Linux,
for directly locking a region of bytes.  It remains to be seen if
this method is effective, and more research is necessary.

### Consideration - Go Garbage (Collection)

Similar to the previous consideration, handling any key material
in memory, if not treated with precision, may lead to _garbage_,
or unaddressable but unreclaimed regions of memory that contain
copies of credentials.

I do not have a solution to this yet.

### Consideration - shield-pipe in Go

The current implementation of SHIELD passes the Encryption Key and
IV via the `shield-pipe` environment.  This is available both
in-memory, and from inside of `/proc`.  We will have to
investigate whether grabbing local copies of exported environment
variables keeps them out of `/proc` or not.  In any event,
`shield-pipe` _must not_ propagate key material to the plugins it
executes, and must therefore scrub its environment.

One solution is to replace the `shield-pipe` bash implementation
for one in Go, and manage memory according to the previous two
Considerations (i.e. don't generate garbage, use `mlock(2)`).
While actionable, such an effort is not trivial.
