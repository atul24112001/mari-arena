"use client";
import { useAuth } from "@/context/auth";
import { useWallet } from "@solana/wallet-adapter-react";
import { useWalletModal } from "@solana/wallet-adapter-react-ui";
import React from "react";

export default function NavbarActions() {
  const wallet = useWallet();
  const walletModal = useWalletModal();
  const { user } = useAuth();

  return (
    <>
      {wallet.connected ? (
        <div className="flex gap-2 justify-between items-center">
          <div className="text-black">
            {(user?.solanaBalance || 0) / 10 ** 9} SOL
          </div>
        </div>
      ) : (
        <button
          className="bg-[#2b9245] text-[#fff] text-xs md:text-sm font-bold  rounded-md px-5 py-2 mb-2"
          onClick={() => walletModal.setVisible(true)}
        >
          Connect
        </button>
      )}
    </>
  );
}
