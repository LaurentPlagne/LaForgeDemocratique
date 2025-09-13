import streamlit as st

st.title("Article Entry Form")

with st.form("article_form"):
    title = st.text_input("Title")
    author = st.text_input("Author")
    content = st.text_area("Content")

    submitted = st.form_submit_button("Submit")
    if submitted:
        st.write("## Entered Article")
        st.write(f"**Title:** {title}")
        st.write(f"**Author:** {author}")
        st.write(f"**Content:** {content}")
