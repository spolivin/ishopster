"""SQLAlchemy ORM models for the product catalog.

Scope: `categories` and `products` only. Relationship is many-to-one
(a product belongs to exactly one category; a category has many products).
"""

from __future__ import annotations

from datetime import datetime
from decimal import Decimal

from sqlalchemy import DateTime, ForeignKey, MetaData, Numeric, String, Text, func
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship

# Explicit naming convention: without it PostgreSQL/SQLAlchemy invent
# names for indexes and constraints, and Alembic autogenerate produces
# unstable migration code. With it, every constraint has a predictable
# name, so migrations stay reversible and diffable.
NAMING_CONVENTION = {
    "ix": "ix_%(column_0_label)s",
    "uq": "uq_%(table_name)s_%(column_0_name)s",
    "ck": "ck_%(table_name)s_%(constraint_name)s",
    "fk": "fk_%(table_name)s_%(column_0_name)s_%(referred_table_name)s",
    "pk": "pk_%(table_name)s",
}


class Base(DeclarativeBase):
    """Declarative base. `Base.metadata` is what Alembic diffs against."""

    metadata = MetaData(naming_convention=NAMING_CONVENTION)


class TimestampMixin:
    """`created_at` / `updated_at`, both DB-managed.

    `server_default=func.now()` makes PostgreSQL fill the value, so rows
    inserted by raw SQL or seed scripts are also stamped. `onupdate` is
    applied by SQLAlchemy on ORM UPDATEs (a raw SQL UPDATE would not
    refresh it -- a DB trigger would, but that is overkill for now).
    """

    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now()
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )


class Category(Base, TimestampMixin):
    __tablename__ = "categories"

    id: Mapped[int] = mapped_column(primary_key=True)
    # Human-readable label shown in the UI.
    name: Mapped[str] = mapped_column(String(255))
    # URL segment: /categories/<slug>. UNIQUE because it identifies the
    # category in a link; a duplicate slug would make a URL ambiguous.
    slug: Mapped[str] = mapped_column(String(255), unique=True)
    description: Mapped[str | None] = mapped_column(Text)

    products: Mapped[list[Product]] = relationship(back_populates="category")


class Product(Base, TimestampMixin):
    __tablename__ = "products"

    id: Mapped[int] = mapped_column(primary_key=True)

    # NOT NULL: every product lives in a category, so a category page can
    # always list it. ondelete="RESTRICT": PostgreSQL refuses to delete a
    # category that still has products (you must move or delete them
    # first). CASCADE would silently delete products with the category;
    # SET NULL would need the column to be nullable.
    # index=True: the core catalog query is "products WHERE category_id =
    # ?"; PostgreSQL does not index foreign keys automatically.
    category_id: Mapped[int] = mapped_column(
        ForeignKey("categories.id", ondelete="RESTRICT"), index=True
    )

    name: Mapped[str] = mapped_column(String(255))
    slug: Mapped[str] = mapped_column(String(255), unique=True)
    description: Mapped[str | None] = mapped_column(Text)

    # Numeric(10, 2), not Float: money must be exact. A binary float
    # cannot represent 0.10 precisely, so sums drift by fractions of a
    # cent. Maps to Python Decimal. 10 digits total, 2 after the point
    # -> up to 99_999_999.99.
    price: Mapped[Decimal] = mapped_column(Numeric(10, 2))

    category: Mapped[Category] = relationship(back_populates="products")
