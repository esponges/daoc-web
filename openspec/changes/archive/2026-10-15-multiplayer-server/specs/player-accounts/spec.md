## ADDED Requirements

### Requirement: Login by Discord or admin password
The server SHALL authenticate a player either through Discord OAuth2 or, when an admin password is configured, through that password, and SHALL refuse every other connection to the world.

#### Scenario: Discord login
- **WHEN** a player completes the Discord OAuth2 flow
- **THEN** the server creates or finds the account for that Discord user id and starts a session

#### Scenario: Admin password login
- **WHEN** the server is started with an admin password and a client submits it
- **THEN** the client gets an administrator session

#### Scenario: Wrong password
- **WHEN** a client submits a wrong admin password
- **THEN** no session is created and repeated failures from one address are rate limited

#### Scenario: No credentials
- **WHEN** a client opens the world WebSocket without a valid session
- **THEN** the server closes it without sending any world data

### Requirement: Secrets stay out of the repository
The server SHALL read the admin password hash, Discord client id and secret, and session key from the environment or a gitignored file, and SHALL NOT start with a default admin password.

#### Scenario: Missing admin password
- **WHEN** the server starts with no admin password and no Discord credentials configured
- **THEN** it exits with a message saying how to configure one

### Requirement: Characters belong to accounts and persist
An account SHALL own characters, each with a name unique across the server, a race, a class and a last position, and the server SHALL restore a character exactly as it was saved.

#### Scenario: Create a character
- **WHEN** a logged-in player creates a Troll Warrior named Grimm
- **THEN** the character exists on the account, starts at its race's start location, and the name is refused to everyone else

#### Scenario: Name taken
- **WHEN** a player creates a character with a name already in use, ignoring case
- **THEN** the server refuses it and creates nothing

#### Scenario: Return after logout
- **WHEN** a player logs out at a position and logs in again
- **THEN** the character appears at that position with the same state

### Requirement: Administrators are marked
Administrator sessions SHALL be able to use game-master commands, and ordinary sessions SHALL NOT.

#### Scenario: Ordinary player uses an admin command
- **WHEN** a non-administrator sends a game-master command such as teleporting another player
- **THEN** the server refuses it and tells the sender

### Requirement: Login failures say why
The server SHALL log the reason a Discord login failed, without logging the secret or the code.

#### Scenario: A rejected token exchange
- **WHEN** Discord refuses the token exchange
- **THEN** the server log names the status Discord returned and the player is shown a generic failure
