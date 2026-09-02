"""Read-only endpoints for the product resource."""

from typing import Annotated

from fastapi import APIRouter, Depends, HTTPException, Query, status
from sqlalchemy import func, select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from ..db import get_db
from ..models import Category, Product
from ..schemas import ProductList, ProductRead

router = APIRouter(prefix="/api/products", tags=["products"])


@router.get("", response_model=ProductList)
async def list_products(
    session: Annotated[AsyncSession, Depends(get_db)],
    offset: Annotated[int, Query(ge=0)] = 0,
    limit: Annotated[int, Query(ge=1, le=100)] = 20,
    category: str | None = None,
) -> ProductList:
    """Return one page of products, optionally filtered by category slug."""
    query = (
        select(Product)
        .options(selectinload(Product.category))
        # name is not unique, so add id as a tie-breaker: without a stable
        # total ordering, OFFSET can skip or repeat rows across pages.
        .order_by(Product.name, Product.id)
        .offset(offset)
        .limit(limit)
    )
    count_query = select(func.count()).select_from(Product)
    if category is not None:
        query = query.join(Category).where(Category.slug == category)
        count_query = count_query.join(Category).where(Category.slug == category)

    items = (await session.execute(query)).scalars().all()
    total = await session.scalar(count_query) or 0

    return ProductList(items=items, total=total, limit=limit, offset=offset)


@router.get("/{slug}", response_model=ProductRead)
async def get_product(
    slug: str, session: Annotated[AsyncSession, Depends(get_db)]
) -> Product:
    """Return a single product by its slug, or 404 if none exists."""
    result = await session.execute(
        select(Product)
        .options(selectinload(Product.category))
        .where(Product.slug == slug)
    )
    product = result.scalar_one_or_none()
    if product is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="product not found"
        )
    return product
