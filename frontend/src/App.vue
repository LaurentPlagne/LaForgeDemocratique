<script setup>
import { ref, onMounted } from 'vue'
const laws = ref([])
const error = ref(null)
onMounted(async () => {
  try {
    const response = await fetch('http://localhost:8080/api/laws')
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    laws.value = await response.json()
  } catch (e) {
    error.value = e.message
    console.error("Failed to fetch laws:", e)
  }
})
</script>
<template>
  <header><h1>Projets de Loi Collaboratifs</h1></header>
  <main>
    <div v-if="error" class="error">
      <h2>Erreur de communication avec le serveur</h2>
      <p>Impossible de charger les données. Avez-vous bien démarré le serveur backend ?</p>
      <pre>{{ error }}</pre>
    </div>
    <div v-else-if="laws.length > 0">
      <h2>Liste des projets de loi</h2>
      <ul><li v-for="law in laws" :key="law.id">{{ law.title }}</li></ul>
    </div>
    <div v-else><p>Chargement des lois...</p></div>
  </main>
</template>
<style>
:root { font-family: system-ui, sans-serif; line-height: 1.5; color-scheme: light dark; color: rgba(255, 255, 255, 0.87); background-color: #242424; }
body { margin: 0; display: flex; place-items: center; min-height: 100vh; }
#app { max-width: 1280px; margin: 0 auto; padding: 2rem; text-align: center; }
header { margin-bottom: 2rem; }
ul { list-style: none; padding: 0; }
li { background-color: #3a3a3a; padding: 1rem; margin-bottom: 0.5rem; border-radius: 8px; }
.error { color: #ff6b6b; border: 1px solid #ff6b6b; background-color: #4d2a2a; padding: 1rem; border-radius: 8px; }
</style>
