-- CreateTable
CREATE TABLE "online_statuses" (
    "user_id" INTEGER NOT NULL,
    "username" VARCHAR(16) NOT NULL,
    "custom_status" VARCHAR(128),
    "is_private" BOOLEAN NOT NULL DEFAULT false,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "online_statuses_pkey" PRIMARY KEY ("user_id")
);
