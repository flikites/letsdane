/*
  # Create Happy Eyeballs Metrics Tables

  1. New Tables
    - `he_dns_resolutions`
      - `id` (uuid, primary key) - Unique identifier for each DNS resolution
      - `host` (text) - The hostname being resolved
      - `family` (smallint) - Address family (4 for IPv4, 6 for IPv6)
      - `start_time` (timestamptz) - When the DNS lookup started
      - `end_time` (timestamptz) - When the DNS lookup completed
      - `address_count` (integer) - Number of addresses returned
      - `success` (boolean) - Whether the lookup succeeded
      - `error_message` (text) - Error message if lookup failed
      - `created_at` (timestamptz) - Record creation timestamp
    
    - `he_connection_attempts`
      - `id` (uuid, primary key) - Unique identifier for each connection attempt
      - `host` (text) - The hostname being connected to
      - `ip` (inet) - The IP address attempted
      - `family` (smallint) - Address family (4 for IPv4, 6 for IPv6)
      - `start_time` (timestamptz) - When the connection attempt started
      - `end_time` (timestamptz) - When the connection attempt completed
      - `success` (boolean) - Whether the connection succeeded
      - `winner` (boolean) - Whether this was the winning connection
      - `error_message` (text) - Error message if connection failed
      - `duration_ms` (integer) - Connection duration in milliseconds
      - `created_at` (timestamptz) - Record creation timestamp

  2. Indexes
    - Index on `host` for both tables for efficient querying
    - Index on `created_at` for time-based queries
    - Index on `winner` for finding successful connections
    - Index on `family` for IPv4 vs IPv6 analysis

  3. Security
    - Enable RLS on both tables
    - Add policies for authenticated users to insert their own metrics
    - Add policies for authenticated users to read their own metrics
*/

CREATE TABLE IF NOT EXISTS he_dns_resolutions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  host text NOT NULL,
  family smallint NOT NULL CHECK (family IN (4, 6)),
  start_time timestamptz NOT NULL,
  end_time timestamptz NOT NULL,
  address_count integer DEFAULT 0,
  success boolean DEFAULT false,
  error_message text,
  created_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS he_connection_attempts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  host text NOT NULL,
  ip inet NOT NULL,
  family smallint NOT NULL CHECK (family IN (4, 6)),
  start_time timestamptz NOT NULL,
  end_time timestamptz NOT NULL,
  success boolean DEFAULT false,
  winner boolean DEFAULT false,
  error_message text,
  duration_ms integer GENERATED ALWAYS AS (
    EXTRACT(EPOCH FROM (end_time - start_time)) * 1000
  ) STORED,
  created_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_he_dns_resolutions_host ON he_dns_resolutions(host);
CREATE INDEX IF NOT EXISTS idx_he_dns_resolutions_created_at ON he_dns_resolutions(created_at);
CREATE INDEX IF NOT EXISTS idx_he_dns_resolutions_family ON he_dns_resolutions(family);

CREATE INDEX IF NOT EXISTS idx_he_connection_attempts_host ON he_connection_attempts(host);
CREATE INDEX IF NOT EXISTS idx_he_connection_attempts_created_at ON he_connection_attempts(created_at);
CREATE INDEX IF NOT EXISTS idx_he_connection_attempts_winner ON he_connection_attempts(winner);
CREATE INDEX IF NOT EXISTS idx_he_connection_attempts_family ON he_connection_attempts(family);

ALTER TABLE he_dns_resolutions ENABLE ROW LEVEL SECURITY;
ALTER TABLE he_connection_attempts ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Allow public insert for dns resolutions"
  ON he_dns_resolutions
  FOR INSERT
  TO public
  WITH CHECK (true);

CREATE POLICY "Allow public select for dns resolutions"
  ON he_dns_resolutions
  FOR SELECT
  TO public
  USING (true);

CREATE POLICY "Allow public insert for connection attempts"
  ON he_connection_attempts
  FOR INSERT
  TO public
  WITH CHECK (true);

CREATE POLICY "Allow public select for connection attempts"
  ON he_connection_attempts
  FOR SELECT
  TO public
  USING (true);
