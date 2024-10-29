/*
  Warnings:

  - A unique constraint covering the columns `[title,currency]` on the table `gametypes` will be added. If there are existing duplicate values, this will fail.

*/
-- CreateIndex
CREATE UNIQUE INDEX "gametypes_title_currency_key" ON "gametypes"("title", "currency");
