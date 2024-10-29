/*
  Warnings:

  - You are about to drop the column `type` on the `games` table. All the data in the column will be lost.
  - Added the required column `gameTypeId` to the `games` table without a default value. This is not possible if the table is not empty.

*/
-- CreateEnum
CREATE TYPE "Currency" AS ENUM ('INR', 'SOL');

-- AlterTable
ALTER TABLE "games" DROP COLUMN "type",
ADD COLUMN     "gameTypeId" TEXT NOT NULL;

-- DropEnum
DROP TYPE "GameType";

-- CreateTable
CREATE TABLE "gametypes" (
    "id" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "entry" INTEGER NOT NULL,
    "winner" INTEGER NOT NULL,
    "currency" "Currency" NOT NULL,

    CONSTRAINT "gametypes_pkey" PRIMARY KEY ("id")
);

-- AddForeignKey
ALTER TABLE "games" ADD CONSTRAINT "games_gameTypeId_fkey" FOREIGN KEY ("gameTypeId") REFERENCES "gametypes"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
