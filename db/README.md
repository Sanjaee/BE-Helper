# Database Configuration

## PostgreSQL & pgAdmin Setup

### Quick Start

1. **Reset and start services (IMPORTANT - Run this first time):**
   ```bash
   # Windows
   reset-db.bat
   
   # Linux/Mac
   chmod +x reset-db.sh
   ./reset-db.sh
   ```

   **OR manually:**
   ```bash
   docker-compose down
   docker volume rm be-helper_postgres_data
   docker volume rm be-helper_pgadmin_data
   docker-compose up -d
   ```

2. **Access pgAdmin:**
   - URL: http://localhost:5050
   - Email: admin@admin.com
   - Password: admin123

3. **Connect to PostgreSQL:**
   - The PostgreSQL server should already be configured in pgAdmin
   - If not, add a new server with these details:
     - Host: postgres (or localhost if connecting from outside Docker)
     - Port: 5432
     - Username: postgres
     - Password: 123
     - Database: postgres

### Available Databases

The following databases are automatically created:

- `postgres` - Default PostgreSQL database
- `userdb` - User service database
- `productdb` - Product service database  
- `paymentdb` - Payment service database

### Troubleshooting

#### pgAdmin shows no databases

1. **Check if PostgreSQL is running:**
   ```bash
   docker-compose ps postgres
   ```

2. **Check PostgreSQL logs:**
   ```bash
   docker-compose logs postgres
   ```

3. **Verify database creation:**
   ```bash
   docker-compose exec postgres psql -U postgres -c "\l"
   ```

4. **Restart services:**
   ```bash
   docker-compose down
   docker-compose up -d
   ```

#### Manual Database Connection in pgAdmin

If the automatic server configuration doesn't work:

1. Right-click "Servers" in pgAdmin
2. Select "Register" → "Server"
3. Fill in the connection details:
   - **General Tab:**
     - Name: PostgreSQL Server
   - **Connection Tab:**
     - Host name/address: `postgres` (or `localhost`)
     - Port: `5432`
     - Username: `postgres`
     - Password: `123`
     - Save password: ✓

### Database Initialization

The database initialization scripts are located in `db/init/`:

- `01_create_databases.sql` - Creates all required databases
- `pgadmin_servers.json` - pgAdmin server configuration

These scripts run automatically when the PostgreSQL container starts for the first time.
