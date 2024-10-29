"use client";

import { IconButton, Modal } from "@/components/helper";
import { useAuth } from "@/context/auth";
import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { useWalletModal } from "@solana/wallet-adapter-react-ui";
import {
  LAMPORTS_PER_SOL,
  PublicKey,
  SystemProgram,
  Transaction,
} from "@solana/web3.js";
import { useMutation } from "@tanstack/react-query";
import axios from "axios";
import { LogOut, Plus } from "lucide-react";
import React, { useEffect, useRef, useState } from "react";

export default function ClientHomePage() {
  const rechargeAmountRef = useRef<HTMLInputElement>(null);

  const [openRechargeModal, setOpenRechargeModal] = useState(false);
  const { user, setUser, apiClient } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [signature, setSignature] = useState<null | string>(null);

  const { mutate, isPending } = useMutation({
    mutationFn: async (signature: string) => {
      const { data } = await apiClient.post(`/api/transaction`, {
        signature,
      });
      return data;
    },
    mutationKey: ["transaction"],
  });

  const wallet = useWallet();
  const walletModal = useWalletModal();
  const { connection } = useConnection();

  const verifySignature = (sign: string) => {
    setLoading(true);
    if (!signature) {
      setSignature(sign);
    }
    const retry = parseInt(localStorage.getItem("verification-retry") || "0");
    setTimeout(() => {
      mutate(sign, {
        onError: (e) => {
          if (retry < 5) {
            localStorage.setItem("verification-retry", `${retry + 1}`);
            verifySignature(sign);
          } else {
            setError(e.message);
            setLoading(false);
          }
        },
        onSuccess: (data) => {
          setUser(data.data[0]);
          toggleRechargeModal();
          setSignature(null);
          setLoading(false);
        },
        onSettled: () => {
          localStorage.removeItem("verification-retry");
        },
      });
    }, 2000);
  };

  const addBalanceHandler = async () => {
    try {
      setLoading(true);
      const amount = parseFloat(rechargeAmountRef.current?.value || "");
      if (!user || isNaN(amount) || !wallet.publicKey || !wallet.connected) {
        console.log("Something went wrong");
        return;
      }

      const to = process.env.NEXT_PUBLIC_KEY as string;
      console.log({ amount, to, fromPubkey: wallet.publicKey });
      const transaction = new Transaction();
      transaction.add(
        SystemProgram.transfer({
          fromPubkey: wallet.publicKey,
          toPubkey: new PublicKey(to),
          lamports: amount * LAMPORTS_PER_SOL,
        })
      );

      const signature = await wallet.sendTransaction(transaction, connection);
      verifySignature(signature);
    } catch (error) {
      if (error instanceof Error) {
        setLoading(false);
        setError(error.message);
      }
    }
  };
  const toggleRechargeModal = () => {
    if (!loading) {
      setOpenRechargeModal((prev) => !prev);
    }
  };

  return (
    <>
      {wallet.connected ? (
        <div className="flex justify-between items-center  gap-2">
          <div className="text-black">
            {(user?.solanaBalance || 0) / 10 ** 9} SOL
          </div>
          <IconButton onClick={toggleRechargeModal}>
            <Plus size={18} />
            <span className="hidden md:inline-block">Add Sol</span>
          </IconButton>
          <IconButton color="error" onClick={wallet.disconnect}>
            <LogOut size={20} />
            <span className="hidden md:inline-block">Disconnect</span>
          </IconButton>
        </div>
      ) : (
        <button
          className="bg-[#2b9245] text-[#fff] text-xs md:text-sm font-bold  rounded-md px-5 py-2 mb-2"
          onClick={() => walletModal.setVisible(true)}
        >
          Connect
        </button>
      )}
      <Modal
        title="Add Solana"
        onClose={toggleRechargeModal}
        open={openRechargeModal}
        className="mt-2"
      >
        {error && (
          <div className="mb-3 text-sm font-bold bg-red-100 text-red-500 rounded-lg px-4 py-1">
            {error}
          </div>
        )}
        <div className="mb-3 text-sm font-bold bg-gray-200 text-gray-500 rounded-lg px-4 py-1">
          Note: In some cases it takes sometime for a transaction to get
          distributed all over blockchain so verification might fail. You have
          to retry the verification in those case.
        </div>
        <input
          ref={rechargeAmountRef}
          placeholder="Amount (2 SOL)"
          type="number"
          className="w-full  rounded-lg  bg-[#2b924520] px-4 py-2 my-2 active:outline-none focus:outline-none"
        />
        <div className="flex justify-end items-center gap-3">
          <IconButton color="error" onClick={toggleRechargeModal}>
            Cancel
          </IconButton>
          {signature ? (
            <IconButton
              loading={loading || isPending}
              disabled={loading}
              onClick={() => verifySignature(signature)}
            >
              Retry
            </IconButton>
          ) : (
            <IconButton
              loading={loading}
              disabled={loading}
              onClick={addBalanceHandler}
            >
              Submit
            </IconButton>
          )}
        </div>
      </Modal>
    </>
  );
}
