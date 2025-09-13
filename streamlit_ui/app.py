import streamlit as st
import requests
import json

# --- Configuration ---
API_BASE_URL = "http://localhost:8080/api"

# --- API Communication ---

def get_laws():
    """Fetches the list of all laws from the backend."""
    try:
        response = requests.get(f"{API_BASE_URL}/laws/")
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        st.error(f"Error fetching laws: {e}")
        return []

def create_new_law(law_title, article_title, article_content, author):
    """Posts a new law and its first article to the backend."""
    url = f"{API_BASE_URL}/laws/"
    payload = {
        "title": law_title,
        "article_title": article_title,
        "article_content": article_content,
        "author": author
    }
    try:
        response = requests.post(url, data=json.dumps(payload), headers={'Content-Type': 'application/json'})
        response.raise_for_status()
        st.success(f"Successfully created new law: '{law_title}'")
        return True
    except requests.exceptions.RequestException as e:
        st.error(f"Error creating law: {e.body}")
        return False

def add_article_to_law(law_id, article_title, article_content, author):
    """Posts a new article to an existing law."""
    url = f"{API_BASE_URL}/laws/{law_id}/articles"
    payload = {
        "title": article_title,
        "content": article_content,
        "author": author
    }
    try:
        response = requests.post(url, data=json.dumps(payload), headers={'Content-Type': 'application/json'})
        response.raise_for_status()
        st.success(f"Successfully added article '{article_title}' to the law.")
        return True
    except requests.exceptions.RequestException as e:
        st.error(f"Error adding article: {e}")
        return False

# --- UI Components ---

st.set_page_config(layout="centered")
st.title("⚖️ Law and Article Management")

laws = get_laws()
law_options = {law['title']: law['id'] for law in laws}

CREATE_NEW_LAW_OPTION = "--- Create a New Law ---"
options_list = [CREATE_NEW_LAW_OPTION] + list(law_options.keys())

st.sidebar.header("Action")
selection = st.sidebar.selectbox("Choose a Law or create a new one:", options_list)

if selection == CREATE_NEW_LAW_OPTION:
    st.header("Create a New Law")
    with st.form("new_law_form"):
        law_title = st.text_input("Law Title")
        st.markdown("---")
        article_title = st.text_input("First Article Title")
        author = st.text_input("Author")
        article_content = st.text_area("First Article Content", height=300)

        submitted = st.form_submit_button("Create Law")
        if submitted:
            if all([law_title, article_title, author, article_content]):
                create_new_law(law_title, article_title, article_content, author)
            else:
                st.warning("Please fill out all fields.")

else:
    st.header(f"Add an Article to '{selection}'")
    law_id = law_options.get(selection)

    with st.form("new_article_form"):
        article_title = st.text_input("Article Title")
        author = st.text_input("Author")
        article_content = st.text_area("Article Content", height=300)

        submitted = st.form_submit_button("Add Article")
        if submitted:
            if all([law_id, article_title, author, article_content]):
                add_article_to_law(law_id, article_title, article_content, author)
            else:
                st.warning("Please fill out all fields.")

st.sidebar.markdown("---")
if st.sidebar.button("Refresh Law List"):
    st.experimental_rerun()
