<template>
  <v-container>
    <a href="http://localhost:8080/web" target="_blank">New User</a>
    <div v-if="!isLogin">
      <v-text-field label="Name" v-model="userName"></v-text-field>
      <v-text-field label="RoomID" v-model="roomID"></v-text-field>
      <v-btn block @click="onJoinBtnClick"> join </v-btn>
    </div>
    <div v-else>
      <v-btn block @click="onPublishBtnClick"> Publish </v-btn>
      <h2>Local Video</h2>
      <user-video
        :id="userId"
        :src="localStream"
        :mute="true"
        :control="true"
      />

      <h2>Remotes Video</h2>
      <user-video
        v-for="(stream, key) in users"
        :id="key"
        :key="key"
        :src="stream"
        :mute="false"
        :control="false"
      />
    </div>
  </v-container>
</template>

<script>
import UserVideo from "../components/UserVideo";
import SFU from "../classes/Webrtc/SFU";

export default {
  name: "Home",
  data: () => {
    return {
      sfu: null,
      connected: false,
      published: false,
      isLogin: false,
      userName: "",
      roomID: "",
      userId: "",
      localStream: null,
      users: {},
    };
  },
  methods: {
    getRandomUserId() {
      var num = "";
      for (var i = 0; i < 6; i++) {
        num += Math.floor(Math.random() * 10);
      }
      return num;
    },
    async onPublishBtnClick() {
      if (!this.connected) {
        console.log("Клиент не подключен к серверу");
        return;
      }
      if (this.published) {
        console.log("Подготовка аудио и видео...");
        return;
      }
      console.log("Раздача потоков");
      this.sfu.publish();
      this.published = true;
    },
    onJoinBtnClick() {
      let v = this;

      if (v.userName == "" || v.roomID == "") {
        alert("Введите корректные значения полей");
        return;
      }

      v.userId = v.getRandomUserId();
      v.sfu = new SFU(v.userId, v.roomID);
      v.sfu.on("connect", () => {
        console.log("Соединение SFU!");
        v.connected = true;
        v.isLogin = true;

        v.sfu.join(v.userId, v.userName, v.roomID);
      });

      v.sfu.on("disconnect", () => {
        v.connected = false;
      });

      v.sfu.on("addLocalStream", (id, stream) => {
        // var user = new User({ id, stream, parent: "localVideoDiv" });
        // v.users.set(id, stream);
        console.log("addLocalStream: ",id);
        console.log(stream);
        v.localStream = stream;
      });

      this.sfu.on("addRemoteStream", (id, stream) => {
        // var user = new User({ id, stream, parent: "remoteVideoDiv" });
        console.log(`Remote userID ${id} connect`);
        v.$set(v.users, id, stream);
      });

      this.sfu.on("removeRemoteStream", (id, stream) => {
        v.$delete(v.users, id)
        console.log(stream);
      });
    },
  },
  mounted: () => {},
  components: {
    UserVideo,
  },
};
</script>
