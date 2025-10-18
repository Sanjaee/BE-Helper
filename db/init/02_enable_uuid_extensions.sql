-- Enable uuid-ossp extension for all databases
-- This script runs after databases are created

-- Enable extension in main postgres database
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable extension in userdb
\c userdb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable extension in orderdb  
\c orderdb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable extension in locationdb
\c locationdb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable extension in paymentdb
\c paymentdb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable extension in notificationdb
\c notificationdb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Switch back to default database
\c postgres;
