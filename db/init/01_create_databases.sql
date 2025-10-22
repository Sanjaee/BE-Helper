-- Create databases
CREATE DATABASE userdb;
CREATE DATABASE orderdb;
CREATE DATABASE locationdb;
CREATE DATABASE paymentdb;
CREATE DATABASE notificationdb;
CREATE DATABASE chatdb;

-- Enable uuid-ossp extension for all databases
\c userdb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c orderdb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c locationdb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c paymentdb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c notificationdb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c chatdb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Switch back to default database
\c postgres;