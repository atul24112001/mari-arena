/*
  Warnings:

  - You are about to drop the `_GameToUser` table. If the table is not empty, all the data it contains will be lost.
  - Added the required column `maxPlayer` to the `games` table without a default value. This is not possible if the table is not empty.

*/
-- DropForeignKey
ALTER TABLE "_GameToUser" DROP CONSTRAINT "_GameToUser_A_fkey";

-- DropForeignKey
ALTER TABLE "_GameToUser" DROP CONSTRAINT "_GameToUser_B_fkey";

-- AlterTable
ALTER TABLE "games" ADD COLUMN     "maxPlayer" INTEGER NOT NULL,
ADD COLUMN     "winnerId" TEXT,
ALTER COLUMN "status" SET DEFAULT 'staging';

-- AlterTable
ALTER TABLE "users" ADD COLUMN     "password" TEXT;

-- DropTable
DROP TABLE "_GameToUser";

-- CreateTable
CREATE TABLE "participants" (
    "gameId" TEXT NOT NULL,
    "userId" TEXT NOT NULL,

    CONSTRAINT "participants_pkey" PRIMARY KEY ("gameId","userId")
);

-- AddForeignKey
ALTER TABLE "participants" ADD CONSTRAINT "participants_gameId_fkey" FOREIGN KEY ("gameId") REFERENCES "games"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "participants" ADD CONSTRAINT "participants_userId_fkey" FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
