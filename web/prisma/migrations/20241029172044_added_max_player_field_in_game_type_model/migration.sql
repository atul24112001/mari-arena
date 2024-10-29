/*
  Warnings:

  - Added the required column `maxPlayer` to the `gametypes` table without a default value. This is not possible if the table is not empty.

*/
-- AlterTable
ALTER TABLE "gametypes" ADD COLUMN     "maxPlayer" INTEGER NOT NULL;
