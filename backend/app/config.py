"""Application configuration, loaded from the environment / `.env`."""

from pathlib import Path

from pydantic_settings import BaseSettings, SettingsConfigDict
from sqlalchemy import URL

# Repo-root .env, resolved from this file's location so it is found no
# matter the working directory (app run from backend/, alembic, tests).
# In the container the file is absent; compose injects the vars directly
# and pydantic-settings simply skips a missing env_file.
ENV_FILE = Path(__file__).resolve().parents[2] / ".env"


class Settings(BaseSettings):
    """Postgres connection settings, populated from the environment."""

    model_config = SettingsConfigDict(env_file=ENV_FILE, extra="ignore")

    postgres_user: str
    postgres_password: str
    postgres_db: str
    postgres_host: str = "localhost"
    postgres_port: int = 5432

    @property
    def database_url(self) -> URL:
        """Assemble the psycopg SQLAlchemy connection URL.

        Built via ``URL.create`` so that special characters in the
        password are escaped correctly.
        """
        return URL.create(
            drivername="postgresql+psycopg",
            username=self.postgres_user,
            password=self.postgres_password,
            host=self.postgres_host,
            port=self.postgres_port,
            database=self.postgres_db,
        )


settings = Settings()
