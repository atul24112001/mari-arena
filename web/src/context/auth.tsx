"use client";
import { IconButton } from "@/components/helper";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { ShootingStars } from "@/components/ui/shooting-stars";
import { StarsBackground } from "@/components/ui/star-background";
import { User } from "@prisma/client";
import { useWallet } from "@solana/wallet-adapter-react";
import { useMutation } from "@tanstack/react-query";
import axios, { AxiosInstance } from "axios";

import {
  Dispatch,
  PropsWithChildren,
  SetStateAction,
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { toast } from "sonner";

type AuthContextType = {
  user: User | null;
  setUser: Dispatch<SetStateAction<User | null>>;
  apiClient: AxiosInstance;
  isAdmin: boolean;
  sendMessage: (type: string, data: any) => void;
  socket: WebSocket | null;
  openPasswordDialog: boolean;
  togglePasswordDialog: () => void;
};

const AuthContext = createContext<AuthContextType>({
  user: null,
  setUser: () => {},
  apiClient: axios,
  isAdmin: false,
  sendMessage: () => {},
  socket: null,
  openPasswordDialog: false,
  togglePasswordDialog: () => {},
});

async function verifyUser(data: {
  name: string;
  identifier: string;
  token: string;
}): Promise<{
  data: [User];
  message: string;
  token: string;
  isAdmin?: boolean;
}> {
  const response = await axios.post(
    `${process.env.NEXT_PUBLIC_API_URL}/api/user`,
    {
      name: data.name,
      identifier: data.identifier,
    },
    {
      headers: {
        Authorization: `Bearer ${data.token}`,
      },
    }
  );
  return response.data;
}

export const AuthContextProvider = ({ children }: PropsWithChildren) => {
  const passwordRef = useRef<HTMLInputElement>(null);

  const [user, setUser] = useState<null | User>(null);
  const [token, setToken] = useState<string | null>(null);

  const [isAdmin, setIsAdmin] = useState(false);
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [openPasswordDialog, setOpenPasswordDialog] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const wallet = useWallet();

  const { mutate } = useMutation({
    mutationFn: verifyUser,
    mutationKey: ["verifyUser"],
  });

  const authenticateMutant = useMutation({
    mutationFn: async (data: { password: string; identifier: string }) => {
      const response = await axios.post(
        `${process.env.NEXT_PUBLIC_API_URL}/api/auth`,
        {
          password: data.password,
          identifier: data.identifier,
        }
      );
      return response.data;
    },
    mutationKey: ["authenticate"],
  });

  const apiClient = useMemo(() => {
    const headers: any = {};
    console.log("apiClient", { token });
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    return axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_URL,
      headers,
    });
  }, [token]);

  useEffect(() => {
    if (user && !socket) {
      const ws = new WebSocket(`${process.env.NEXT_PUBLIC_WSS}/ws`);

      ws.onopen = () => {
        setSocket(ws);
      };

      ws.onclose = (e) => {
        console.log(e);
        wallet.disconnect();
        setUser(null);
        setToken(null);
        localStorage.removeItem("token");
      };

      ws.onerror = (e) => {
        console.log(e);
        wallet.disconnect();
        setUser(null);
        setToken(null);
        localStorage.removeItem("token");
      };

      return () => {
        ws.close();
      };
    }
  }, [user, socket]);

  const sendMessage = useCallback(
    (type: string, data: any) => {
      if (socket) {
        socket.send(
          JSON.stringify({
            type,
            data,
          })
        );
      }
    },
    [socket]
  );

  useEffect(() => {
    if (socket && user) {
      sendMessage("add-user", {
        userId: user.id,
        publicKey: user.email,
      });

      socket.onmessage = (e) => {
        const { type } = JSON.parse(e.data);
        if (type === "refresh") {
          window.location.reload();
        }
      };
    }
  }, [socket, user]);

  const togglePasswordDialog = () => setOpenPasswordDialog((prev) => !prev);

  useEffect(() => {
    const _token = localStorage.getItem("token");
    const publicKey = wallet.publicKey?.toBase58();
    if (wallet.connected) {
      if (_token && publicKey) {
        mutate(
          {
            name: publicKey,
            identifier: publicKey,
            token: _token,
          },
          {
            onError: () => {
              wallet.disconnect();
            },
            onSuccess: (data) => {
              setUser(data.data[0]);
              setToken(_token);
              if (data.isAdmin) {
                setIsAdmin(true);
              }
              setOpenPasswordDialog(false);
            },
          }
        );
      } else {
        setUser(null);
        setOpenPasswordDialog(true);
      }
    }
  }, [wallet.connected]);

  const authenticate = () => {
    if (!wallet.publicKey) {
      toast("Please connect a wallet first");
      togglePasswordDialog();
      return;
    }
    const password = passwordRef.current?.value || "";
    if (password.length > 15) {
      toast("Password length should be less then 15");
      return;
    }
    if (password.length < 7) {
      toast("Password length should be more then 7");
      return;
    }
    authenticateMutant.mutate(
      {
        identifier: wallet.publicKey?.toBase58(),
        password,
      },
      {
        onSuccess: (data) => {
          togglePasswordDialog();
          setUser(data.data);
          setToken(data.token);
          localStorage.setItem("token", data.token);
        },
        onError: (error) => {
          // @ts-ignore
          toast(error?.response?.data?.message || error.message);
        },
      }
    );
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        sendMessage,
        setUser,
        apiClient,
        isAdmin,
        socket,
        openPasswordDialog,
        togglePasswordDialog,
      }}
    >
      <div className="h-screen dark bg-[#000] text-white  relative overflow-hidden bg-cover md:bg-contain">
        {openPasswordDialog ? (
          <div className="text-white h-screen relative z-10 flex justify-center items-center">
            <div>
              <Input
                ref={passwordRef}
                className="text-black"
                placeholder="Password"
                type={showPassword ? "text" : "password"}
              />
              <div className="cursor-pointer flex items-center space-x-2 mt-2">
                <Checkbox
                  checked={showPassword}
                  onCheckedChange={(v) => setShowPassword(v as boolean)}
                  id="terms"
                  className="dark"
                />
                <label
                  htmlFor="terms"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  Show Password
                </label>
              </div>

              <div className="flex justify-center mt-3">
                <IconButton onClick={authenticate}>Unlock</IconButton>
              </div>
            </div>
          </div>
        ) : (
          children
        )}

        <StarsBackground />
        <ShootingStars />
      </div>
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
