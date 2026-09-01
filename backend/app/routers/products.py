"""Read-only endpoints for the product resource."""

from typing import Annotated

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from ..db import get_db
from ..models import Category, Product
from ..schemas import ProductRead

router = APIRouter(prefix="/api/products", tags=["products"])


@router.get("", response_model=list[ProductRead])
async def list_products(
    session: Annotated[AsyncSession, Depends(get_db)],
    category: str | None = None,
):
    """Return products, optionally filtered by category slug."""
    query = select(Product).order_by(Product.name)
    if category is not None:
        query = query.join(Category).where(Category.slug == category)
    result = await session.execute(query)
    return result.scalars().all()


@router.get("/{slug}", response_model=ProductRead)
async def get_product(slug: str, session: Annotated[AsyncSession, Depends(get_db)]):
    """Return a single product by its slug, or 404 if none exists."""
    result = await session.execute(select(Product).where(Product.slug == slug))
    product = result.scalar_one_or_none()
    if product is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="product not found"
        )
    return product
