"use client";
import { useAuth } from "@/context/auth";
import React, { PropsWithChildren } from "react";

export default function Admin({ children }: PropsWithChildren) {
  const { isAdmin } = useAuth();
  if (!isAdmin) {
    return null;
  }
  return <>{children}</>;
}
