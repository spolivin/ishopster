"""Read-only endpoints for the category resource."""

from typing import Annotated

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from ..db import get_db
from ..models import Category
from ..schemas import CategoryRead

router = APIRouter(prefix="/api/categories", tags=["categories"])


@router.get("", response_model=list[CategoryRead])
async def list_categories(session: Annotated[AsyncSession, Depends(get_db)]):
    """Return every category, ordered by name."""
    result = await session.execute(select(Category).order_by(Category.name))
    return result.scalars().all()


@router.get("/{slug}", response_model=CategoryRead)
async def get_category(slug: str, session: Annotated[AsyncSession, Depends(get_db)]):
    """Return a single category by its slug, or 404 if none exists."""
    result = await session.execute(select(Category).where(Category.slug == slug))
    category = result.scalar_one_or_none()
    if category is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="category not found"
        )
    return category
