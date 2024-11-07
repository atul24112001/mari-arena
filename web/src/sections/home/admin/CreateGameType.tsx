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
import { AxiosError } from "axios";
import React, { useEffect, useState } from "react";
import { toast } from "sonner";

export default function AdminCreateGameType() {
  const { apiClient, isAdmin, underMaintenance, toggleUnderMaintenance } =
    useAuth();
  const [openGameTypeForm, setOpenGameTypeForm] = useState(false);
  const [counter, setCounter] = useState({
    activeUsers: 0,
    ongoingGames: 0,
  });
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

  useEffect(() => {
    if (isAdmin) {
      try {
        (async () => {
          const { data } = await apiClient.get("/api/admin/metric");
          setCounter({
            activeUsers: data.data.activeUsers.length,
            ongoingGames: data.data.ongoingGames.length,
          });
        })();
      } catch (error) {
        console.log(error);
      }
    }
  }, [isAdmin]);

  const handleAddGameType = () => {
    mutate(details, {
      onSuccess: () => {
        window.location.reload();
      },
    });
  };

  const toggleMaintenance = async () => {
    try {
      await apiClient.get("/api/admin/maintenance");
      toggleUnderMaintenance();
    } catch (error) {
      if (error instanceof AxiosError) {
        toast(error.response?.data?.message || error.message);
      }
    }
  };

  return (
    <Admin>
      <div className="text-primary flex flex-col items-center justify-center">
        <IconButton onClick={toggleGameTypeForm}>Add Game Type</IconButton>
        <IconButton onClick={toggleMaintenance}>
          {underMaintenance ? "End" : "Start"} maintenance
        </IconButton>
        <IconButton>{counter.activeUsers} active users</IconButton>
        <IconButton>{counter.ongoingGames} active games</IconButton>
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
