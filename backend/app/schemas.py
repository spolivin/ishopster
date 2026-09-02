from decimal import Decimal

from pydantic import BaseModel, ConfigDict


class CategoryRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)
    id: int
    name: str
    slug: str
    description: str | None = None


class ProductRead(BaseModel):
    model_config = ConfigDict(from_attributes=True)
    id: int
    name: str
    slug: str
    description: str | None = None
    price: Decimal
    category: CategoryRead


class ProductList(BaseModel):
    """One page of products plus the totals a client needs to paginate."""

    items: list[ProductRead]
    total: int  # total matching products, ignoring limit/offset
    limit: int
    offset: int
