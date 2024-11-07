"use client";
import { IconButton } from "@/components/helper";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { ShootingStars } from "@/components/ui/shooting-stars";
import { StarsBackground } from "@/components/ui/star-background";
import { User } from "@prisma/client";
import { useWallet } from "@solana/wallet-adapter-react";
import { useMutation } from "@tanstack/react-query";
import axios, { AxiosError, AxiosInstance } from "axios";

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
  underMaintenance: boolean;
  togglePasswordDialog: () => void;
  toggleUnderMaintenance: () => void;
};

const AuthContext = createContext<AuthContextType>({
  user: null,
  setUser: () => {},
  apiClient: axios,
  isAdmin: false,
  sendMessage: () => {},
  socket: null,
  openPasswordDialog: false,
  underMaintenance: false,
  togglePasswordDialog: () => {},
  toggleUnderMaintenance: () => {},
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
  underMaintenance?: boolean;
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
  const [underMaintenance, setUnderMaintenance] = useState(false);
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [openPasswordDialog, setOpenPasswordDialog] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [retry, setRetry] = useState(0);

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
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    return axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_URL,
      headers,
    });
  }, [token]);

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

  const connectSocket = () => {
    setRetry((prev) => prev + 1);
    const ws = new WebSocket(`${process.env.NEXT_PUBLIC_WSS}`);
    ws.onopen = () => {
      setSocket(ws);
    };
    ws.onclose = () => {
      if (retry < 6) {
        setSocket(null);
        connectSocket();
      }
      // console.log(JSON.stringify(e));
      // wallet.disconnect();
      // setUser(null);
      // setToken(null);
      // localStorage.removeItem("token");
    };

    ws.onerror = () => {
      if (retry < 6) {
        setSocket(null);
        connectSocket();
      }
    };
  };

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
            onError: (e) => {
              wallet.disconnect();
              if (e instanceof AxiosError) {
                toast(e.response?.data.message);
              }
            },
            onSuccess: (data) => {
              connectSocket();
              setUser(data.data[0]);
              setToken(_token);
              if (data.isAdmin) {
                setIsAdmin(true);
              }
              if (data.underMaintenance) {
                setUnderMaintenance(true);
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
          connectSocket();
          setUser(data.data);
          setToken(data.token);
          localStorage.setItem("token", data.token);
        },
        onError: (e) => {
          if (e instanceof AxiosError) {
            toast(e.response?.data.message);
          }
        },
      }
    );
  };

  const toggleUnderMaintenance = () => setUnderMaintenance((prev) => !prev);

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
        toggleUnderMaintenance,
        underMaintenance,
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
