<template>
  <v-container>
    <h2 v-if="source == null">
      Нет потокового видео
    </h2>
    <video ref="video" autoplay :muted="{mute}" :controls="{control}" v-else></video>
  </v-container>
</template>

<script>
  export default {
    name: 'UserVideo',
    props: [ "id", "src", "mute","control",],
    data: () => ({
      source: null
    }),
    methods: {
      async loadSource(val) {
        console.log(val);
        let v = this;
        v.source = val;
        setTimeout(()=>{
          v.$refs.video.srcObject = val;
          clearTimeout(this);
        }, 400);
      }
    },
    created() {
      if (this.src != null) {
        this.loadSource(this.src)
      }
    },
    watch: {
      src: {
         handler: function(val) {
            this.loadSource(val); 
        },
        deep: true,
      }
    }
  }
</script>

<style scoped>
video {
  width:320px;
  height:240px;
}
</style>