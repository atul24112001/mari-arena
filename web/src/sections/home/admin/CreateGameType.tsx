"use client";
import { IconButton, Modal } from "@/components/helper";
import Admin from "@/components/helper/Admin";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuth } from "@/context/auth";
import { GameType } from "@prisma/client";
import { useMutation } from "@tanstack/react-query";
import React, { useState } from "react";

export default function AdminCreateGameType() {
  const { apiClient } = useAuth();
  const [openGameTypeForm, setOpenGameTypeForm] = useState(false);
  const [details, setDetails] = useState<GameType>({
    currency: "INR",
    entry: 0,
    id: "",
    title: "",
    winner: 0,
    maxPlayer: 2,
  });

  const { mutate, error } = useMutation({
    mutationFn: async (payload: GameType) => {
      await apiClient.post("/api/game-types", {
        title: payload.title,
        entry: parseInt(`${payload.entry}`),
        winner: parseInt(`${payload.winner}`),
        maxPlayer: parseInt(`${payload.maxPlayer}`),
        currency: payload.currency,
      });
    },
  });

  const toggleGameTypeForm = () => setOpenGameTypeForm((prev) => !prev);

  const changeHandler = (name: string, value: string | number) => {
    setDetails((prev) => {
      return {
        ...prev,
        [name]: value,
      };
    });
  };

  const handleAddGameType = () => {
    mutate(details, {
      onSuccess: () => {
        window.location.reload();
      },
    });
  };

  return (
    <Admin>
      <div className="text-primary flex justify-center">
        <IconButton onClick={toggleGameTypeForm}>Add Game Type</IconButton>
      </div>
      <Modal
        onClose={toggleGameTypeForm}
        open={openGameTypeForm}
        title="Add game type"
        className="my-2 gap-3 flex flex-col text-black"
        // @ts-ignore
        error={error?.response?.data.message || error?.message || ""}
      >
        <Input
          value={details.title}
          onChange={(e) => changeHandler("title", e.target.value)}
          name="title"
          placeholder="Title"
        />
        <Input
          type="number"
          value={details.entry}
          onChange={(e) => changeHandler("entry", e.target.value)}
          name="entry"
          placeholder="Entry"
        />
        <Input
          type="number"
          value={details.maxPlayer}
          onChange={(e) => changeHandler("maxPlayer", e.target.value)}
          name="maxPlayer"
          placeholder="Max Player"
        />
        <Input
          type="number"
          onChange={(e) => changeHandler("winner", e.target.value)}
          value={details.winner}
          name="winner"
          placeholder="Winner"
        />
        <Select onValueChange={(value) => changeHandler("currency", value)}>
          <SelectTrigger className="w-full">
            <SelectValue placeholder="Select a currency" />
          </SelectTrigger>
          <SelectContent onChange={() => {}}>
            <SelectGroup>
              <SelectLabel>Currency</SelectLabel>
              <SelectItem value="INR">INR</SelectItem>
              <SelectItem value="SOL">SOL</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <div className="flex justify-end items-center gap-2">
          <IconButton color="error" onClick={toggleGameTypeForm}>
            Cancel
          </IconButton>
          <IconButton onClick={handleAddGameType}>Save</IconButton>
        </div>
      </Modal>
    </Admin>
  );
}
